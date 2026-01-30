package deps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HomebrewDependency represents the Homebrew package manager
type HomebrewDependency struct{}

// NewHomebrewDependency creates a new Homebrew dependency
func NewHomebrewDependency() *HomebrewDependency {
	return &HomebrewDependency{}
}

// Name returns the dependency name
func (h *HomebrewDependency) Name() string {
	return "Homebrew"
}

// Check returns the current status of Homebrew
func (h *HomebrewDependency) Check(ctx context.Context) (Status, error) {
	if !HasBrew() {
		return StatusNotInstalled, nil
	}
	return StatusInstalled, nil
}

// Install installs Homebrew using the official installation script
func (h *HomebrewDependency) Install(ctx context.Context) error {
	platform := DetectPlatform()
	if !platform.IsDarwin() {
		return fmt.Errorf("homebrew installation is only supported on macOS")
	}

	// Run the official Homebrew installation script
	installScript := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Homebrew: %w", err)
	}

	// After installation, update PATH for the current process
	if err := h.updatePath(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	return nil
}

// Start is a no-op for Homebrew (doesn't need to be started)
func (h *HomebrewDependency) Start(ctx context.Context) error {
	return nil
}

// RequiresStart returns false as Homebrew doesn't need to be started
func (h *HomebrewDependency) RequiresStart() bool {
	return false
}

// updatePath adds Homebrew to the current process's PATH
func (h *HomebrewDependency) updatePath() error {
	brewPath := GetBrewPath()

	currentPath := os.Getenv("PATH")
	if strings.Contains(currentPath, brewPath) {
		return nil
	}

	newPath := brewPath + ":" + currentPath
	if err := os.Setenv("PATH", newPath); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	return nil
}

// GetVersion returns the installed Homebrew version
func (h *HomebrewDependency) GetVersion(ctx context.Context) (string, error) {
	output, err := GetCommandVersion(ctx, "brew", "--version")
	if err != nil {
		return "", err
	}

	// Output format: "Homebrew 4.x.x"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}

	return output, nil
}
