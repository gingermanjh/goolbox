package deps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// FFmpegDependency represents the ffmpeg video processing tool
type FFmpegDependency struct{}

// NewFFmpegDependency creates a new ffmpeg dependency
func NewFFmpegDependency() *FFmpegDependency {
	return &FFmpegDependency{}
}

// Name returns the dependency name
func (f *FFmpegDependency) Name() string {
	return "ffmpeg"
}

// Check returns the current status of ffmpeg
func (f *FFmpegDependency) Check(ctx context.Context) (Status, error) {
	if !CommandExists("ffmpeg") {
		return StatusNotInstalled, nil
	}
	return StatusInstalled, nil
}

// Install installs ffmpeg using the platform-specific installer
func (f *FFmpegDependency) Install(ctx context.Context) error {
	platform := DetectPlatform()

	switch {
	case platform.IsDarwin():
		return f.installDarwin(ctx)
	case platform.IsLinux():
		return f.installLinux(ctx)
	default:
		return fmt.Errorf("ffmpeg installation is not supported on %s", platform.OS)
	}
}

// Start is a no-op for ffmpeg (doesn't need to be started)
func (f *FFmpegDependency) Start(ctx context.Context) error {
	return nil
}

// RequiresStart returns false as ffmpeg doesn't need to be started
func (f *FFmpegDependency) RequiresStart() bool {
	return false
}

// installDarwin installs ffmpeg on macOS using Homebrew
func (f *FFmpegDependency) installDarwin(ctx context.Context) error {
	if !HasBrew() {
		return fmt.Errorf("homebrew is required to install ffmpeg on macOS")
	}

	cmd := exec.CommandContext(ctx, "brew", "install", "ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install ffmpeg via Homebrew: %w", err)
	}

	return nil
}

// installLinux installs ffmpeg on Linux
func (f *FFmpegDependency) installLinux(ctx context.Context) error {
	// Try apt-get first (Debian/Ubuntu)
	if CommandExists("apt-get") {
		cmd := exec.CommandContext(ctx, "sudo", "apt-get", "install", "-y", "ffmpeg")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try dnf (Fedora)
	if CommandExists("dnf") {
		cmd := exec.CommandContext(ctx, "sudo", "dnf", "install", "-y", "ffmpeg")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try yum (CentOS/RHEL)
	if CommandExists("yum") {
		cmd := exec.CommandContext(ctx, "sudo", "yum", "install", "-y", "ffmpeg")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not find a supported package manager to install ffmpeg")
}

// GetVersion returns the installed ffmpeg version
func (f *FFmpegDependency) GetVersion(ctx context.Context) (string, error) {
	output, err := GetCommandVersion(ctx, "ffmpeg", "-version")
	if err != nil {
		return "", err
	}

	// Output format: "ffmpeg version 6.x.x Copyright..."
	// Extract version number
	re := regexp.MustCompile(`ffmpeg version (\S+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// Fallback: return first line
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}

	return output, nil
}
