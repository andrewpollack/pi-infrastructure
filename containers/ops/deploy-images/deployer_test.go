package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

// FakeRunner records executed commands.
type FakeRunner struct {
	Commands   []string
	ShouldFail bool
}

func (fr *FakeRunner) Run(name string, args ...string) error {
	if fr.ShouldFail {
		return fmt.Errorf("simulated failure")
	}
	fr.Commands = append(fr.Commands, fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
	return nil
}

func (fr *FakeRunner) RunSilent(name string, args ...string) error {
	if fr.ShouldFail {
		return fmt.Errorf("simulated failure")
	}
	fr.Commands = append(fr.Commands, fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
	return nil
}

// Fake implementations for dependency injection.
func fakeGetLocalImageID(img Image) (string, error) {
	// For the test image, return a matching ID.
	if img.Name == "app" && img.Tag == "latest" && img.Registry == "registry.example.com" {
		return "sha256:match", nil
	}
	return "", fmt.Errorf("no fake local image id for %s", img.RegistryString())
}

func fakeGetRemoteImageID(img Image, tgt Target) (string, error) {
	// Return "sha256:match" if target Host is "remote-match"; else "sha256:different".
	if img.Name == "app" && img.Tag == "latest" && img.Registry == "registry.example.com" {
		if tgt.Host == "remote-match" {
			return "sha256:match", nil
		}
		return "sha256:different", nil
	}
	return "", nil
}

func fakeGetRemoteFileHashUpToDate(tgt Target, remoteFile string) (string, error) {
	// Compute the local hash for the compose file so that it matches.
	return computeLocalFileHash(tgt.LocalCompose)
}

func fakeGetRemoteFileHashOutdated(tgt Target, remoteFile string) (string, error) {
	// Return a hash that is different.
	return "differenthash", nil
}

func createTempFile(content string, prefix string) (string, error) {
	tmpFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			log.Printf("failed to close temp file: %v", err)
		}
	}()

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func TestComputeLocalFileHash(t *testing.T) {
	content := "hello world"
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	hash, err := computeLocalFileHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("computeLocalFileHash failed: %v", err)
	}
	expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if hash != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, hash)
	}
}

func TestCheckImagesUpToDate(t *testing.T) {
	images := []Image{
		{Name: "app", Tag: "latest", Registry: "registry.example.com"},
	}
	target := Target{Host: "remote-match", TargetPath: "/tmp"}
	deployer := &Deployer{
		GetLocalImageIDFunc:  fakeGetLocalImageID,
		GetRemoteImageIDFunc: fakeGetRemoteImageID,
	}
	upToDate, err := deployer.CheckImagesUpToDate(target, images)
	if err != nil {
		t.Fatalf("CheckImagesUpToDate error: %v", err)
	}
	if !upToDate {
		t.Errorf("expected images to be up to date")
	}
	target.Host = "remote-different"
	upToDate, err = deployer.CheckImagesUpToDate(target, images)
	if err != nil {
		t.Fatalf("CheckImagesUpToDate error: %v", err)
	}
	if upToDate {
		t.Errorf("expected images to be not up to date")
	}
}

func TestCheckComposeUpToDate(t *testing.T) {
	content := "version: '3'\nservices:\n  app:\n    image: registry.example.com/app:latest\n"
	composePath, err := createTempFile(content, "compose*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp compose file: %v", err)
	}
	defer func() {
		if err := os.Remove(composePath); err != nil {
			log.Printf("failed to remove temp compose file: %v", err)
		}
	}()

	target := Target{
		Host:         "remote",
		LocalCompose: composePath,
		TargetPath:   "/remote/path",
	}
	deployer := &Deployer{
		GetRemoteFileHashFunc: fakeGetRemoteFileHashUpToDate,
	}
	upToDate, err := deployer.CheckComposeUpToDate(target)
	if err != nil {
		t.Fatalf("CheckComposeUpToDate error: %v", err)
	}
	if !upToDate {
		t.Errorf("expected compose to be up to date")
	}
}

func TestDeployTarget_SkipDeployment(t *testing.T) {
	content := "version: '3'\nservices:\n  app:\n    image: registry.example.com/app:latest\n"
	composePath, err := createTempFile(content, "compose*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp compose file: %v", err)
	}
	defer func() {
		if err := os.Remove(composePath); err != nil {
			log.Printf("failed to remove temp compose file: %v", err)
		}
	}()

	configContent := "test config"
	configPath, err := createTempFile(configContent, "config*.json")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer func() {
		if err := os.Remove(configPath); err != nil {
			log.Printf("failed to remove temp config file: %v", err)
		}
	}()

	target := Target{
		Host:         "remote-match",
		LocalCompose: composePath,
		LocalConfig:  configPath,
		TargetPath:   "/remote/path",
	}
	images := []Image{
		{Name: "app", Tag: "latest", Registry: "registry.example.com"},
	}
	fr := &FakeRunner{
		ShouldFail: true,
	}
	deployer := &Deployer{
		Runner:                fr,
		GetLocalImageIDFunc:   fakeGetLocalImageID,
		GetRemoteImageIDFunc:  fakeGetRemoteImageID,
		GetRemoteFileHashFunc: fakeGetRemoteFileHashUpToDate,
	}
	err = deployer.DeployTarget(target, images)
	if err == nil {
		t.Fatalf("DeployTarget should have failed")
	}
	// Expect no commands because images and compose are up to date.
	if len(fr.Commands) != 0 {
		t.Errorf("expected no commands when deployment is skipped, got: %v", fr.Commands)
	}
}

func TestDeployTarget_UpdateImages(t *testing.T) {
	content := "version: '3'\nservices:\n  app:\n    image: registry.example.com/app:latest\n"
	composePath, err := createTempFile(content, "compose*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp compose file: %v", err)
	}
	defer func() {
		if err := os.Remove(composePath); err != nil {
			log.Printf("failed to remove temp compose file: %v", err)
		}
	}()

	configContent := "test config"
	configPath, err := createTempFile(configContent, "config*.json")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer func() {
		if err := os.Remove(configPath); err != nil {
			log.Printf("failed to remove temp config file: %v", err)
		}
	}()

	target := Target{
		Host:         "remote-different",
		LocalCompose: composePath,
		LocalConfig:  configPath,
		TargetPath:   "/remote/path",
	}
	images := []Image{
		{Name: "app", Tag: "latest", Registry: "registry.example.com"},
	}
	fr := &FakeRunner{
		Commands: []string{}, // Ensure commands are reset
	}
	deployer := &Deployer{
		Runner:                fr,
		GetLocalImageIDFunc:   fakeGetLocalImageID,
		GetRemoteImageIDFunc:  fakeGetRemoteImageID,
		GetRemoteFileHashFunc: fakeGetRemoteFileHashUpToDate,
	}
	err = deployer.DeployTarget(target, images)
	if err != nil {
		t.Fatalf("DeployTarget failed: %v", err)
	}
	// Expect commands to update images
	if len(fr.Commands) == 0 {
		t.Errorf("expected commands to update images, got none")
	}
}

func TestDeployTarget_UpdateCompose(t *testing.T) {
	content := "version: '3'\nservices:\n  app:\n    image: registry.example.com/app:latest\n"
	composePath, err := createTempFile(content, "compose*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp compose file: %v", err)
	}
	defer func() {
		if err := os.Remove(composePath); err != nil {
			log.Printf("failed to remove temp compose file: %v", err)
		}
	}()

	configContent := "test config"
	configPath, err := createTempFile(configContent, "config*.json")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer func() {
		if err := os.Remove(configPath); err != nil {
			log.Printf("failed to remove temp config file: %v", err)
		}
	}()

	target := Target{
		Host:         "remote-match",
		LocalCompose: composePath,
		LocalConfig:  configPath,
		TargetPath:   "/remote/path",
	}
	images := []Image{
		{Name: "app", Tag: "latest", Registry: "registry.example.com"},
	}
	fr := &FakeRunner{
		Commands: []string{}, // Ensure commands are reset
	}
	deployer := &Deployer{
		Runner:                fr,
		GetLocalImageIDFunc:   fakeGetLocalImageID,
		GetRemoteImageIDFunc:  fakeGetRemoteImageID,
		GetRemoteFileHashFunc: fakeGetRemoteFileHashOutdated,
	}
	err = deployer.DeployTarget(target, images)
	if err != nil {
		t.Fatalf("DeployTarget failed: %v", err)
	}
	// Expect commands to update compose file
	if len(fr.Commands) == 0 {
		t.Errorf("expected commands to update compose file, got none")
	}
}

func TestDeployTarget_UpdateBoth(t *testing.T) {
	content := "version: '3'\nservices:\n  app:\n    image: registry.example.com/app:latest\n"
	composePath, err := createTempFile(content, "compose*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp compose file: %v", err)
	}
	defer func() {
		if err := os.Remove(composePath); err != nil {
			log.Printf("failed to remove temp compose file: %v", err)
		}
	}()

	configContent := "test config"
	configPath, err := createTempFile(configContent, "config*.json")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer func() {
		if err := os.Remove(configPath); err != nil {
			log.Printf("failed to remove temp config file: %v", err)
		}
	}()

	target := Target{
		Host:         "remote-different",
		LocalCompose: composePath,
		LocalConfig:  configPath,
		TargetPath:   "/remote/path",
	}
	images := []Image{
		{Name: "app", Tag: "latest", Registry: "registry.example.com"},
	}
	fr := &FakeRunner{
		Commands: []string{}, // Ensure commands are reset
	}
	deployer := &Deployer{
		Runner:                fr,
		GetLocalImageIDFunc:   fakeGetLocalImageID,
		GetRemoteImageIDFunc:  fakeGetRemoteImageID,
		GetRemoteFileHashFunc: fakeGetRemoteFileHashOutdated,
	}
	err = deployer.DeployTarget(target, images)
	if err != nil {
		t.Fatalf("DeployTarget failed: %v", err)
	}
	// Expect commands to update both images and compose file
	if len(fr.Commands) == 0 {
		t.Errorf("expected commands to update both images and compose file, got none")
	}
}
