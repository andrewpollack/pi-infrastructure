package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// CommandRunner abstracts running external commands.
type CommandRunner interface {
	Run(name string, args ...string) error
	RunSilent(name string, args ...string) error
}

// DefaultRunner uses os/exec to run commands.
type DefaultRunner struct {
	// Timeout for each external command.
	Timeout time.Duration
}

func (r *DefaultRunner) Run(name string, args ...string) error {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 240 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("command %s timed out after %d", name, timeout)
	}
	return err
}

func (r *DefaultRunner) RunSilent(name string, args ...string) error {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 240 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("command %s timed out", name)
	}
	return err
}

// Config holds targets and images.
type Config struct {
	Targets []Target `json:"targets"`
	Images  []Image  `json:"images"`
}

// LoadConfig reads the JSON config file.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Target represents a remote host.
type Target struct {
	Host         string `json:"host"`
	LocalCompose string `json:"local_compose"`
	LocalConfig  string `json:"local_config"`
	TargetPath   string `json:"target_path"`
}

// Image represents a Docker image.
type Image struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Registry string `json:"registry"`
}

// RegistryString returns "registry/name:tag".
func (i Image) RegistryString() string {
	return fmt.Sprintf("%s/%s:%s", i.Registry, i.Name, i.Tag)
}

// TarString returns the tar filename for the image.
func (i Image) TarString() string {
	return fmt.Sprintf("%s.tar", i.Name)
}

// TarPath returns the full path for the tar file.
func (i Image) TarPath() string {
	return fmt.Sprintf("/tmp/%s", i.TarString())
}

// Deployer performs deployment steps.
type Deployer struct {
	Runner CommandRunner
	// Timeout for all commands run directly from Deployer.
	Timeout time.Duration

	// Function fields to allow dependency injection for testing.
	GetLocalImageIDFunc   func(img Image) (string, error)
	GetRemoteImageIDFunc  func(img Image, tgt Target) (string, error)
	GetRemoteFileHashFunc func(tgt Target, remoteFile string) (string, error)
}

// GetLocalImageID returns the local image ID.
func (d *Deployer) GetLocalImageID(img Image) (string, error) {
	if d.GetLocalImageIDFunc != nil {
		return d.GetLocalImageIDFunc(img)
	}
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.Id}}", img.RegistryString())
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("⏰ command timed out while inspecting local image %s", img.RegistryString())
	}
	if err != nil {
		return "", fmt.Errorf("failed to inspect local image %s: %w", img.RegistryString(), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteImageID returns the remote image ID.
func (d *Deployer) GetRemoteImageID(img Image, tgt Target) (string, error) {
	if d.GetRemoteImageIDFunc != nil {
		return d.GetRemoteImageIDFunc(img, tgt)
	}
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh", tgt.Host, "docker", "inspect", "-f", "{{.Id}}", img.RegistryString())
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("⏰ command timed out while inspecting remote image %s on %s", img.RegistryString(), tgt.Host)
	}
	if err != nil {
		// If the command fails, assume the image is missing remotely.
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteFileHash returns the SHA256 hash of a remote file.
func (d *Deployer) GetRemoteFileHash(tgt Target, remoteFile string) (string, error) {
	if d.GetRemoteFileHashFunc != nil {
		return d.GetRemoteFileHashFunc(tgt, remoteFile)
	}
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ssh", tgt.Host, "sha256sum", remoteFile)
	output, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("⏰ command timed out while computing remote file hash for %s", remoteFile)
	}
	if err != nil {
		return "", fmt.Errorf("failed to compute remote hash for %s: %w", remoteFile, err)
	}
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return "", fmt.Errorf("unexpected output from sha256sum")
	}
	return parts[0], nil
}

// computeLocalFileHash computes a file's SHA256 hash.
func computeLocalFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing file: %v\n", err)
		}
	}()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("failed to compute hash for file %s: %w", filePath, err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// CheckImagesUpToDate returns true if all remote image IDs match the local ones.
func (d *Deployer) CheckImagesUpToDate(tgt Target, images []Image) (bool, error) {
	for _, img := range images {
		localID, err := d.GetLocalImageID(img)
		if err != nil {
			return false, fmt.Errorf("failed to get local image id for %s: %w", img.RegistryString(), err)
		}
		remoteID, err := d.GetRemoteImageID(img, tgt)
		if err != nil {
			return false, fmt.Errorf("failed to get remote image id for %s on %s: %w", img.RegistryString(), tgt.Host, err)
		}
		if localID == "" || localID != remoteID {
			return false, nil
		}
	}
	return true, nil
}

// CheckComposeUpToDate returns true if the remote compose file hash matches the local one.
func (d *Deployer) CheckConfigUpToDate(tgt Target) (bool, error) {
	localHash, err := computeLocalFileHash(tgt.LocalConfig)
	if err != nil {
		return false, fmt.Errorf("failed to compute hash for local compose file %s: %w", tgt.LocalConfig, err)
	}
	configFileName := filepath.Base(tgt.LocalConfig)
	remoteConfigPath := fmt.Sprintf("%s/%s", tgt.TargetPath, configFileName)
	remoteHash, err := d.GetRemoteFileHash(tgt, remoteConfigPath)
	if err != nil {
		// If we can't get the remote hash, assume it's outdated.
		return false, nil
	}
	return localHash == remoteHash, nil
}

// CheckComposeUpToDate returns true if the remote compose file hash matches the local one.
func (d *Deployer) CheckComposeUpToDate(tgt Target) (bool, error) {
	localHash, err := computeLocalFileHash(tgt.LocalCompose)
	if err != nil {
		return false, fmt.Errorf("failed to compute hash for local compose file %s: %w", tgt.LocalCompose, err)
	}
	composeFileName := filepath.Base(tgt.LocalCompose)
	remoteComposePath := fmt.Sprintf("%s/%s", tgt.TargetPath, composeFileName)
	remoteHash, err := d.GetRemoteFileHash(tgt, remoteComposePath)
	if err != nil {
		// If we can't get the remote hash, assume it's outdated.
		return false, nil
	}
	return localHash == remoteHash, nil
}

// PullImage pulls an image.
func (d *Deployer) PullImage(img Image) error {
	if err := d.Runner.RunSilent("docker", "pull", "--platform", "linux/arm64", img.RegistryString()); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", img.RegistryString(), err)
	}
	fmt.Printf("⤵️  Pulled %s\n", img.RegistryString())
	return nil
}

// PackageImage saves an image as a tar file.
func (d *Deployer) PackageImage(img Image) error {
	if err := d.Runner.Run("docker", "save", img.RegistryString(), "-o", img.TarPath()); err != nil {
		return fmt.Errorf("failed to package image %s: %w", img.RegistryString(), err)
	}
	fmt.Printf("📦 Packaged %s\n", img.RegistryString())
	return nil
}

// CopyImage copies the image tar to the remote target.
func (d *Deployer) CopyImage(img Image, tgt Target) error {
	fmt.Printf("Copying %s to %s...\n", img.TarPath(), tgt.Host)
	dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
	if err := d.Runner.Run("scp", img.TarPath(), dst); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", img.TarPath(), dst, err)
	}
	fmt.Printf("Copied %s to %s\n", img.TarPath(), tgt.Host)
	return nil
}

// LoadImage loads the image tar on the remote target.
func (d *Deployer) LoadImage(img Image, tgt Target) error {
	fmt.Printf("Loading %s on %s...\n", img.TarString(), tgt.Host)
	remoteFile := fmt.Sprintf("%s/%s", tgt.TargetPath, img.TarString())
	if err := d.Runner.Run("ssh", tgt.Host, "docker", "load", "-i", remoteFile); err != nil {
		return fmt.Errorf("failed to load %s on %s: %w", img.TarString(), tgt.Host, err)
	}
	fmt.Printf("Loaded %s on %s\n", img.TarString(), tgt.Host)
	return nil
}

// CopyAndLoadImagesForTarget copies and loads images concurrently.
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
			if localID == remoteID && localID != "" {
				fmt.Printf("⏭️  Image %s already present on %s (ID: %s), skipping copy/load\n", img.RegistryString(), tgt.Host, localID)
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
// It skips the compose step only if both images and compose config are up to date.
func (d *Deployer) DeployTarget(tgt Target, images []Image) error {
	fmt.Printf("🔄 Gathering state of %s...\n", tgt.Host)
	imagesUpToDate, err := d.CheckImagesUpToDate(tgt, images)
	if err != nil {
		return err
	}
	composeUpToDate, err := d.CheckComposeUpToDate(tgt)
	if err != nil {
		return err
	}
	configUpToDate, err := d.CheckConfigUpToDate(tgt)
	if err != nil {
		return err
	}

	if imagesUpToDate && composeUpToDate && configUpToDate {
		fmt.Printf("⏭️  No changes on %s (images, compose, and config are up to date), skipping deployment.\n", tgt.Host)
		return nil
	}

	if !imagesUpToDate {
		fmt.Printf("🔄 Copying/loading images on %s...\n", tgt.Host)
		if err := d.CopyAndLoadImagesForTarget(tgt, images); err != nil {
			return fmt.Errorf("deployment failed on %s (copy/load images): %w", tgt.Host, err)
		}
		fmt.Printf("✅ Successfully updated images on %s\n", tgt.Host)
	}

	if !composeUpToDate {
		fmt.Printf("🔄 Copying compose file to %s...\n", tgt.Host)
		dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
		if err := d.Runner.Run("scp", tgt.LocalCompose, dst); err != nil {
			return fmt.Errorf("failed to copy compose file to %s: %w", tgt.Host, err)
		}
		fmt.Printf("Copied compose file to %s\n", tgt.Host)
	}

	if !configUpToDate {
		fmt.Printf("🔄 Copying config file to %s...\n", tgt.Host)
		dst := fmt.Sprintf("%s:%s", tgt.Host, tgt.TargetPath)
		if err := d.Runner.Run("scp", tgt.LocalConfig, dst); err != nil {
			return fmt.Errorf("failed to copy config file to %s: %w", tgt.Host, err)
		}
		fmt.Printf("Copied config file to %s\n", tgt.Host)
	}

	composeFileName := filepath.Base(tgt.LocalCompose)

	fmt.Printf("🔄 Running compose on %s with file %s...\n", tgt.Host, composeFileName)
	if err := d.Runner.Run("ssh", tgt.Host,
		"cd", tgt.TargetPath, "&&",
		"docker", "compose", "-f", composeFileName, "down", "&&",
		"docker", "compose", "-f", composeFileName, "up", "-d"); err != nil {
		return fmt.Errorf("failed to run compose on %s: %w", tgt.Host, err)
	}
	fmt.Printf("✅ Successfully ran compose on %s\n", tgt.Host)
	return nil
}

// DeployTargets deploys concurrently to all targets.
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

func (d *Deployer) PullImages(images []Image) []error {
	return runConcurrent(images, func(img Image) error {
		return d.PullImage(img)
	})
}

func (d *Deployer) PackageImages(images []Image) []error {
	return runConcurrent(images, func(img Image) error {
		return d.PackageImage(img)
	})
}

func main() {
	rfs, err := runfiles.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize runfiles: %v\n", err)
		os.Exit(1)
	}

	// TODO: Is there a better way to find the config file?
	path, err := rfs.Rlocation("_main/containers/ops/deploy-images/data/config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to resolve runfile path: %v\n", err)
		os.Exit(1)
	}

	configPath := flag.String("config", path, "Path to configuration file")
	timeoutFlag := flag.Duration("timeout", 240*time.Second, "Timeout for external commands")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Set the timeout for both the Runner and the Deployer.
	runner := &DefaultRunner{Timeout: *timeoutFlag}
	deployer := Deployer{Runner: runner, Timeout: *timeoutFlag}

	fmt.Println("🔄 Pulling images...")
	if errs := deployer.PullImages(config.Images); errs != nil {
		fmt.Println("❌ Errors occurred while pulling images:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All images successfully pulled.")
	fmt.Println("🔄 Packaging images...")
	if errs := deployer.PackageImages(config.Images); errs != nil {
		fmt.Println("❌ Errors occurred while packaging images:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All images packaged.")
	fmt.Println("🔄 Deploying to all targets (copy/load + compose in parallel)...")
	if errs := deployer.DeployTargets(config.Targets, config.Images); errs != nil {
		fmt.Println("❌ Errors occurred during deployment:")
		for _, e := range errs {
			fmt.Println(" -", e)
		}
		os.Exit(1)
	}
	fmt.Println("✅ All targets deployed successfully.")
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
