package deps

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SupportedContainerMacOSMajor is the minimum macOS major version supported by
// Apple's container runtime.
const SupportedContainerMacOSMajor = 26

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

// IsAppleSilicon returns true if running on Apple Silicon (ARM64 Mac)
func (p Platform) IsAppleSilicon() bool {
	return p.OS == "darwin" && p.Arch == "arm64"
}

// SupportsContainer reports whether the host can run Apple's container runtime.
// It requires macOS 26 or later on Apple silicon.
func SupportsContainer() bool {
	p := DetectPlatform()
	return p.IsDarwin() && p.IsAppleSilicon() && MacOSMajorVersion() >= SupportedContainerMacOSMajor
}

// MacOSMajorVersion returns the major version of the running macOS, or 0 when
// it cannot be determined (for example on other operating systems).
func MacOSMajorVersion() int {
	out, err := RunCommand(context.Background(), "sw_vers", "-productVersion")
	if err != nil {
		return 0
	}

	major, _, _ := strings.Cut(strings.TrimSpace(out), ".")
	n, err := strconv.Atoi(major)
	if err != nil {
		return 0
	}
	return n
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
