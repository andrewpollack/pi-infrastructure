package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type CommandRunner interface {
	Run(name string, args ...string) error
}

type DefaultRunner struct{}

func (r *DefaultRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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

type Deployer struct {
	Runner CommandRunner
}

// PullImage pulls an image (arm64 platform) from the registry.
func (d *Deployer) PullImage(img Image) error {
	err := d.Runner.Run("docker", "pull", "--platform", "linux/arm64", img.RegistryString())
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", img.RegistryString(), err)
	}
	fmt.Printf("Pulled %s\n", img.RegistryString())
	return nil
}

// PackageImage saves the pulled image as a tar file locally.
func (d *Deployer) PackageImage(img Image) error {
	err := d.Runner.Run("docker", "save", img.RegistryString(), "-o", img.TarPath())
	if err != nil {
		return fmt.Errorf("failed to package image %s: %w", img.RegistryString(), err)
	}
	fmt.Printf("Packaged %s\n", img.RegistryString())
	return nil
}

// CopyImage scp's the tar file to the specified target.
func (d *Deployer) CopyImage(img Image, tgt Target) error {
	fmt.Printf("Copying %s to %s...\n", img.TarPath(), tgt.Host)
	dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
	err := d.Runner.Run("scp", img.TarPath(), dst)
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", img.TarPath(), dst, err)
	}
	fmt.Printf("Copied %s to %s\n", img.TarPath(), tgt.Host)
	return nil
}

// LoadImage loads a previously-copied tar file on the remote target (via ssh).
func (d *Deployer) LoadImage(img Image, tgt Target) error {
	fmt.Printf("Loading %s on %s...\n", img.TarString(), tgt.Host)
	remoteFile := fmt.Sprintf("%s/%s", tgt.TargetPath, img.TarString())
	err := d.Runner.Run("ssh", tgt.Host, "docker", "load", "-i", remoteFile)
	if err != nil {
		return fmt.Errorf("failed to load %s on %s: %w", img.TarString(), tgt.Host, err)
	}
	fmt.Printf("Loaded %s on %s\n", img.TarString(), tgt.Host)
	return nil
}

// GetLocalImageID returns the image ID of a local Docker image.
func (d *Deployer) GetLocalImageID(img Image) (string, error) {
	output, err := exec.Command("docker", "inspect", "-f", "{{.Id}}", img.RegistryString()).Output()
	if err != nil {
		return "", fmt.Errorf("failed to inspect local image %s: %w", img.RegistryString(), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteImageID returns the image ID of a Docker image on the remote host.
func (d *Deployer) GetRemoteImageID(img Image, tgt Target) (string, error) {
	// Execute the inspect command on the remote host via ssh.
	cmd := exec.Command("ssh", tgt.Host, "docker", "inspect", "-f", "{{.Id}}", img.RegistryString())
	output, err := cmd.Output()
	if err != nil {
		// If the command fails, assume the image does not exist remotely.
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// Compose copies the local compose file to the target and runs it via Docker Compose.
func (d *Deployer) Compose(tgt Target) error {
	dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
	err := d.Runner.Run("scp", tgt.LocalCompose, dst)
	if err != nil {
		return fmt.Errorf("failed to copy compose file to %s: %w", tgt.Host, err)
	}
	fmt.Printf("Copied compose file to %s\n", tgt.Host)

	err = d.Runner.Run("ssh", tgt.Host,
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

// CopyAndLoadImagesForTarget copies and loads all images for a single target concurrently.
// It first compares the local and remote image IDs. If they match (and the remote image exists),
// the copy and load steps are skipped.
func (d *Deployer) CopyAndLoadImagesForTarget(tgt Target, images []Image) error {
	var wg sync.WaitGroup
	errsCh := make(chan error, len(images))

	for _, img := range images {
		wg.Add(1)
		go func(img Image) {
			defer wg.Done()

			localID, err := d.GetLocalImageID(img)
			if err != nil {
				errsCh <- fmt.Errorf("failed to get local image id for %s: %w", img.RegistryString(), err)
				return
			}
			remoteID, err := d.GetRemoteImageID(img, tgt)
			if err != nil {
				errsCh <- fmt.Errorf("failed to get remote image id for %s on %s: %w", img.RegistryString(), tgt.Host, err)
				return
			}
			// If remote image exists and IDs match, skip copy and load.
			if localID == remoteID && localID != "" {
				fmt.Printf("⏭️ Image %s already present on %s (ID: %s), skipping copy/load\n", img.RegistryString(), tgt.Host, localID)
				return
			}

			if err := d.CopyImage(img, tgt); err != nil {
				errsCh <- err
				return
			}
			if err := d.LoadImage(img, tgt); err != nil {
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
		return fmt.Errorf("copy/load images on target %s encountered errors: %v", tgt.Host, allErrs)
	}
	return nil
}

// DeployTarget performs the entire deployment for a single target.
func (d *Deployer) DeployTarget(tgt Target, images []Image) error {
	fmt.Printf("🔄 Deploying images to %s...\n", tgt.Host)
	if err := d.CopyAndLoadImagesForTarget(tgt, images); err != nil {
		return fmt.Errorf("deployment failed on %s (copy/load images): %w", tgt.Host, err)
	}
	fmt.Printf("✅ Successfully copied and loaded images on %s\n", tgt.Host)

	fmt.Printf("🔄 Running compose on %s...\n", tgt.Host)
	if err := d.Compose(tgt); err != nil {
		return fmt.Errorf("deployment failed on %s (compose): %w", tgt.Host, err)
	}
	fmt.Printf("✅ Successfully ran compose on %s\n", tgt.Host)
	return nil
}

// DeployTargets runs the DeployTarget step in parallel for all targets.
func (d *Deployer) DeployTargets(targets []Target, images []Image) []error {
	var wg sync.WaitGroup
	errsCh := make(chan error, len(targets))

	for _, tgt := range targets {
		wg.Add(1)
		go func(tgt Target) {
			defer wg.Done()
			if err := d.DeployTarget(tgt, images); err != nil {
				errsCh <- err
			}
		}(tgt)
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

// runConcurrent is a helper that runs jobs concurrently.
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
func (d *Deployer) PullImages(images []Image) []error {
	return runConcurrent(images, func(img Image) error {
		return d.PullImage(img)
	})
}

// PackageImages concurrently packages a list of Images.
func (d *Deployer) PackageImages(images []Image) []error {
	return runConcurrent(images, func(img Image) error {
		return d.PackageImage(img)
	})
}

func main() {
	configPath := flag.String("config", "data/config.json", "Path to configuration file")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create a deployer with the default command runner.
	deployer := Deployer{Runner: &DefaultRunner{}}

	// 1. Pull images
	fmt.Println("🔄 Pulling images...")
	if errs := deployer.PullImages(config.Images); errs != nil {
		fmt.Println("Errors occurred while pulling images:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All images successfully pulled.")

	// 2. Package images
	fmt.Println("🔄 Packaging images...")
	if errs := deployer.PackageImages(config.Images); errs != nil {
		fmt.Println("Errors occurred while packaging images:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All images packaged.")

	// 3. Deploy to each target.
	fmt.Println("🔄 Deploying to all targets (copy/load + compose in parallel)...")
	if errs := deployer.DeployTargets(config.Targets, config.Images); errs != nil {
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
