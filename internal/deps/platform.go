package deps

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
)

// Platform represents the current operating system and architecture
type Platform struct {
	OS   string
	Arch string
}

// DetectPlatform returns information about the current platform
func DetectPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// IsDarwin returns true if running on macOS
func (p Platform) IsDarwin() bool {
	return p.OS == "darwin"
}

// IsLinux returns true if running on Linux
func (p Platform) IsLinux() bool {
	return p.OS == "linux"
}

// IsWindows returns true if running on Windows
func (p Platform) IsWindows() bool {
	return p.OS == "windows"
}

// IsAppleSilicon returns true if running on Apple Silicon (ARM64 Mac)
func (p Platform) IsAppleSilicon() bool {
	return p.OS == "darwin" && p.Arch == "arm64"
}

// CommandExists checks if a command exists in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetCommandVersion runs a command with --version or -version flag and returns the output
func GetCommandVersion(ctx context.Context, name string, versionFlag string) (string, error) {
	if versionFlag == "" {
		versionFlag = "--version"
	}

	cmd := exec.CommandContext(ctx, name, versionFlag)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Some commands output version to stderr
		if stderr.Len() > 0 {
			return strings.TrimSpace(stderr.String()), nil
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunCommand runs a command and returns combined output
func RunCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return strings.TrimSpace(output.String()), err
}

// RunCommandSilent runs a command without capturing output
func RunCommandSilent(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

// HasBrew checks if Homebrew is installed
func HasBrew() bool {
	return CommandExists("brew")
}

// GetBrewPath returns the Homebrew binary path based on architecture
func GetBrewPath() string {
	p := DetectPlatform()
	if p.IsAppleSilicon() {
		return "/opt/homebrew/bin"
	}
	return "/usr/local/bin"
}
