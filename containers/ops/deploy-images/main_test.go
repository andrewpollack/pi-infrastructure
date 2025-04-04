package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// FakeRunner is a test double for CommandRunner.
type FakeRunner struct {
	calls []string
	// err to return on every Run call (if set).
	err error
	mu  sync.Mutex
}

func (f *FakeRunner) Run(name string, args ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	call := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	f.calls = append(f.calls, call)
	return f.err
}

// helper to create a new Deployer with a fake runner.
func newTestDeployer(errToReturn error) (*Deployer, *FakeRunner) {
	fr := &FakeRunner{err: errToReturn}
	deployer := &Deployer{Runner: fr}
	return deployer, fr
}

func TestPullImage(t *testing.T) {
	deployer, fr := newTestDeployer(nil)
	img := Image{
		Name:     "nginx",
		Tag:      "latest",
		Registry: "docker.io",
	}
	if err := deployer.PullImage(img); err != nil {
		t.Errorf("PullImage error: %v", err)
	}
	expected := "docker pull --platform linux/arm64 docker.io/nginx:latest"
	if len(fr.calls) != 1 || fr.calls[0] != expected {
		t.Errorf("expected call %q, got %v", expected, fr.calls)
	}
}

func TestPackageImage(t *testing.T) {
	deployer, fr := newTestDeployer(nil)
	img := Image{
		Name:     "nginx",
		Tag:      "latest",
		Registry: "docker.io",
	}
	if err := deployer.PackageImage(img); err != nil {
		t.Errorf("PackageImage error: %v", err)
	}
	expected := "docker save docker.io/nginx:latest -o /tmp/nginx.tar"
	if len(fr.calls) != 1 || fr.calls[0] != expected {
		t.Errorf("expected call %q, got %v", expected, fr.calls)
	}
}

func TestCompose(t *testing.T) {
	deployer, fr := newTestDeployer(nil)
	tgt := Target{
		Host:         "example.com",
		TargetPath:   "/remote",
		LocalCompose: "compose.yml",
	}
	if err := deployer.Compose(tgt); err != nil {
		t.Errorf("Compose error: %v", err)
	}
	if len(fr.calls) != 2 {
		t.Errorf("expected 2 calls, got %d: %v", len(fr.calls), fr.calls)
	}
	expectedScp := "scp compose.yml example.com:/remote"
	expectedSSH := "ssh example.com cd /remote && docker compose down && docker compose up -d"
	if fr.calls[0] != expectedScp {
		t.Errorf("expected first call %q, got %q", expectedScp, fr.calls[0])
	}
	if fr.calls[1] != expectedSSH {
		t.Errorf("expected second call %q, got %q", expectedSSH, fr.calls[1])
	}
}
