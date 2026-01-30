package deps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultDockerWaitTimeout is the default timeout for waiting for Docker daemon
	DefaultDockerWaitTimeout = 120 * time.Second
	// DockerPollInterval is how often to check if Docker daemon is ready
	DockerPollInterval = 2 * time.Second
)

// DockerDependency represents Docker
type DockerDependency struct {
	waitTimeout time.Duration
}

// NewDockerDependency creates a new Docker dependency
func NewDockerDependency() *DockerDependency {
	return &DockerDependency{
		waitTimeout: DefaultDockerWaitTimeout,
	}
}

// WithWaitTimeout sets a custom wait timeout
func (d *DockerDependency) WithWaitTimeout(timeout time.Duration) *DockerDependency {
	d.waitTimeout = timeout
	return d
}

// Name returns the dependency name
func (d *DockerDependency) Name() string {
	return "Docker"
}

// Check returns the current status of Docker
func (d *DockerDependency) Check(ctx context.Context) (Status, error) {
	if !CommandExists("docker") {
		return StatusNotInstalled, nil
	}

	// Check if Docker daemon is running
	if d.isDaemonRunning(ctx) {
		return StatusRunning, nil
	}

	return StatusStopped, nil
}

// Install installs Docker using the platform-specific installer
func (d *DockerDependency) Install(ctx context.Context) error {
	platform := DetectPlatform()

	switch {
	case platform.IsDarwin():
		return d.installDarwin(ctx)
	case platform.IsLinux():
		return d.installLinux(ctx)
	default:
		return fmt.Errorf("docker installation is not supported on %s", platform.OS)
	}
}

// Start starts the Docker daemon
func (d *DockerDependency) Start(ctx context.Context) error {
	platform := DetectPlatform()

	switch {
	case platform.IsDarwin():
		return d.startDarwin(ctx)
	case platform.IsLinux():
		return d.startLinux(ctx)
	default:
		return fmt.Errorf("docker start is not supported on %s", platform.OS)
	}
}

// RequiresStart returns true as Docker needs its daemon running
func (d *DockerDependency) RequiresStart() bool {
	return true
}

// isDaemonRunning checks if the Docker daemon is running
func (d *DockerDependency) isDaemonRunning(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "info")
	return cmd.Run() == nil
}

// installDarwin installs Docker on macOS using Homebrew Cask
func (d *DockerDependency) installDarwin(ctx context.Context) error {
	if !HasBrew() {
		return fmt.Errorf("homebrew is required to install Docker on macOS")
	}

	cmd := exec.CommandContext(ctx, "brew", "install", "--cask", "docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Docker via Homebrew: %w", err)
	}

	return nil
}

// installLinux installs Docker on Linux
func (d *DockerDependency) installLinux(ctx context.Context) error {
	// Use the official Docker installation script
	cmd := exec.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}

	return nil
}

// startDarwin starts Docker Desktop on macOS
func (d *DockerDependency) startDarwin(ctx context.Context) error {
	// Open Docker.app
	cmd := exec.CommandContext(ctx, "open", "-a", "Docker")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Docker Desktop: %w", err)
	}

	// Wait for daemon to be ready
	return d.waitForDocker(ctx)
}

// startLinux starts Docker daemon on Linux
func (d *DockerDependency) startLinux(ctx context.Context) error {
	// Try systemctl first
	if CommandExists("systemctl") {
		cmd := exec.CommandContext(ctx, "sudo", "systemctl", "start", "docker")
		if err := cmd.Run(); err == nil {
			return d.waitForDocker(ctx)
		}
	}

	// Try service command
	if CommandExists("service") {
		cmd := exec.CommandContext(ctx, "sudo", "service", "docker", "start")
		if err := cmd.Run(); err == nil {
			return d.waitForDocker(ctx)
		}
	}

	return fmt.Errorf("could not start Docker daemon. Please start it manually")
}

// waitForDocker waits for the Docker daemon to be ready
func (d *DockerDependency) waitForDocker(ctx context.Context) error {
	timeout := d.waitTimeout
	deadline := time.Now().Add(timeout)

	fmt.Println("Waiting for Docker daemon to be ready...")

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.isDaemonRunning(ctx) {
			return nil
		}

		fmt.Print(".")
		time.Sleep(DockerPollInterval)
	}

	fmt.Println()
	return fmt.Errorf("timeout waiting for Docker daemon to start (waited %v)", timeout)
}

// GetVersion returns the installed Docker version
func (d *DockerDependency) GetVersion(ctx context.Context) (string, error) {
	output, err := GetCommandVersion(ctx, "docker", "--version")
	if err != nil {
		return "", err
	}

	// Output format: "Docker version 24.x.x, build ..."
	re := regexp.MustCompile(`Docker version (\S+),`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// Fallback
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}

	return output, nil
}
