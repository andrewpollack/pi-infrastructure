package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Config struct {
	Targets []Target `json:"targets"`
	Images  []Image  `json:"images"`
}

// LoadConfig reads a JSON config file from the given path.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Target represents a remote host configuration.
type Target struct {
	Host         string `json:"host"`
	LocalCompose string `json:"local_compose"`
	TargetPath   string `json:"target_path"`
}

// Image represents a Docker image to be pulled, packaged, copied, and loaded.
type Image struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Registry string `json:"registry"`
}

// RegistryString returns the "registry/name:tag" format.
func (i Image) RegistryString() string {
	return fmt.Sprintf("%s/%s:%s", i.Registry, i.Name, i.Tag)
}

// TarString returns the local tar filename for the image.
func (i Image) TarString() string {
	return fmt.Sprintf("%s.tar", i.Name)
}

func (i Image) TarPath() string {
	return fmt.Sprintf("/tmp/%s", i.TarString())
}

// runCommand wraps exec.Command for reuse, logging output to stdout/stderr.
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PullImage pulls an image (arm64 platform) from the registry.
func PullImage(img Image) error {
	err := runCommand("docker", "pull", "--platform", "linux/arm64", img.RegistryString())
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", img.RegistryString(), err)
	}
	fmt.Printf("Pulled %s\n", img.RegistryString())
	return nil
}

// PackageImage saves the pulled image as a tar file locally.
func PackageImage(img Image) error {
	err := runCommand("docker", "save", img.RegistryString(), "-o", img.TarPath())
	if err != nil {
		return fmt.Errorf("failed to package image %s: %w", img.RegistryString(), err)
	}
	fmt.Printf("Packaged %s\n", img.RegistryString())
	return nil
}

// CopyImage scp's the tar file to the specified target.
func CopyImage(img Image, tgt Target) error {
	fmt.Printf("Copying %s to %s...\n", img.TarPath(), tgt.Host)
	dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
	err := runCommand("scp", img.TarPath(), dst)
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", img.TarPath(), dst, err)
	}
	fmt.Printf("Copied %s to %s\n", img.TarPath(), tgt.Host)
	return nil
}

// LoadImage loads a previously-copied tar file on the remote target (via ssh).
func LoadImage(img Image, tgt Target) error {
	fmt.Printf("Loading %s on %s...\n", img.TarString(), tgt.Host)
	remoteFile := fmt.Sprintf("%s/%s", tgt.TargetPath, img.TarString())
	err := runCommand("ssh", tgt.Host, "docker", "load", "-i", remoteFile)
	if err != nil {
		return fmt.Errorf("failed to load %s on %s: %w", img.TarString(), tgt.Host, err)
	}
	fmt.Printf("Loaded %s on %s\n", img.TarString(), tgt.Host)
	return nil
}

// CopyAndLoadImagesForTarget copies+loads all images for a single target, but now concurrently.
func CopyAndLoadImagesForTarget(t Target, images []Image) error {
	var wg sync.WaitGroup
	errsCh := make(chan error, len(images))

	for _, img := range images {
		wg.Add(1)
		go func(img Image) {
			defer wg.Done()
			// Copy
			if err := CopyImage(img, t); err != nil {
				errsCh <- err
				return
			}
			// Load
			if err := LoadImage(img, t); err != nil {
				errsCh <- err
			}
		}(img)
	}

	wg.Wait()
	close(errsCh)

	var allErrs []error
	for err := range errsCh {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) > 0 {
		return fmt.Errorf("copy/load images on target %s encountered errors: %v", t.Host, allErrs)
	}
	return nil
}

// Compose copies the local compose file to the target and runs it via Docker Compose.
func Compose(tgt Target) error {
	// Copy the compose file to the target.
	dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
	err := runCommand("scp", tgt.LocalCompose, dst)
	if err != nil {
		return fmt.Errorf("failed to copy compose file to %s: %w", tgt.Host, err)
	}
	fmt.Printf("Copied compose file to %s\n", tgt.Host)

	// Run docker compose on the target.
	err = runCommand("ssh", tgt.Host,
		"cd", tgt.TargetPath, "&&",
		"docker", "compose", "down", "&&",
		"docker", "compose", "up", "-d",
	)
	if err != nil {
		return fmt.Errorf("failed to run compose on %s: %w", tgt.Host, err)
	}
	fmt.Printf("Ran docker compose on %s\n", tgt.Host)
	return nil
}

// DeployTarget performs the entire deployment for a single target:
// 1) Copy and load all images (in parallel) to that target.
// 2) Copy and run the compose file on that target.
func DeployTarget(t Target, images []Image) error {
	// 1. Copy/load images (now concurrently for each image).
	fmt.Printf("🔄 Deploying images to %s...\n", t.Host)
	if err := CopyAndLoadImagesForTarget(t, images); err != nil {
		return fmt.Errorf("deployment failed on %s (copy/load images): %w", t.Host, err)
	}
	fmt.Printf("✅ Successfully copied and loaded images on %s\n", t.Host)
	// 2. Docker compose
	fmt.Printf("🔄 Running compose on %s...\n", t.Host)
	if err := Compose(t); err != nil {
		return fmt.Errorf("deployment failed on %s (compose): %w", t.Host, err)
	}
	fmt.Printf("✅ Successfully ran compose on %s\n", t.Host)
	return nil
}

// DeployTargets runs the DeployTarget step in parallel for all targets.
func DeployTargets(targets []Target, images []Image) []error {
	var wg sync.WaitGroup
	errsCh := make(chan error, len(targets))

	for _, t := range targets {
		wg.Add(1)
		go func(target Target) {
			defer wg.Done()
			if err := DeployTarget(target, images); err != nil {
				errsCh <- err
			}
		}(t)
	}

	wg.Wait()
	close(errsCh)

	var allErrs []error
	for err := range errsCh {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) > 0 {
		return allErrs
	}
	return nil
}

// runConcurrent is useful for “pulling” and “packaging” images in parallel.
func runConcurrent[T any](items []T, job func(T) error) []error {
	var wg sync.WaitGroup
	errsCh := make(chan error, len(items))

	for _, item := range items {
		wg.Add(1)
		go func(it T) {
			defer wg.Done()
			if err := job(it); err != nil {
				errsCh <- err
			}
		}(item)
	}

	wg.Wait()
	close(errsCh)

	var allErrs []error
	for err := range errsCh {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return allErrs
}

// PullImages concurrently pulls a list of Images.
func PullImages(images []Image) []error {
	return runConcurrent(images, PullImage)
}

// PackageImages concurrently packages a list of Images.
func PackageImages(images []Image) []error {
	return runConcurrent(images, PackageImage)
}

func main() {
	configPath := flag.String("config", "data/config.json", "Path to configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 1. Pull images
	fmt.Println("🔄 Pulling images...")
	if errs := PullImages(config.Images); errs != nil {
		fmt.Println("Errors occurred while pulling images:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All images successfully pulled.")

	// 2. Package images
	fmt.Println("🔄 Packaging images...")
	if errs := PackageImages(config.Images); errs != nil {
		fmt.Println("Errors occurred while packaging images:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All images packaged.")

	// 3. Deploy to each target (copy/load images + compose).
	// Each target runs those steps in series, but all targets run in parallel.
	fmt.Println("🔄 Deploying to all targets (copy/load + compose in parallel)...")
	if errs := DeployTargets(config.Targets, config.Images); errs != nil {
		fmt.Println("Errors occurred during deployment:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All targets deployed successfully.")

	// 4. Cleanup local tar files
	fmt.Println("🔄 Cleaning up local tar files...")
	var cleanupErrs []error
	for _, img := range config.Images {
		if err := os.Remove(img.TarPath()); err != nil {
			cleanupErrs = append(cleanupErrs, err)
		}
	}
	if len(cleanupErrs) > 0 {
		fmt.Println("Some tar files could not be removed:")
		for _, e := range cleanupErrs {
			fmt.Println(" -", e)
		}
	} else {
		fmt.Println("✅ Cleanup complete.")
	}

	if len(cleanupErrs) == 0 {
		fmt.Println("✅ All operations completed successfully.")
	} else {
		fmt.Println("Operations completed, but with cleanup issues.")
	}
}
