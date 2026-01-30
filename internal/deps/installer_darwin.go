//go:build darwin

package deps

import (
	"context"
	"fmt"
	"os"
)

// DarwinInstaller provides macOS-specific installation utilities
type DarwinInstaller struct{}

// NewDarwinInstaller creates a new Darwin installer
func NewDarwinInstaller() *DarwinInstaller {
	return &DarwinInstaller{}
}

// EnsureHomebrew ensures Homebrew is installed
func (i *DarwinInstaller) EnsureHomebrew(ctx context.Context, manager *Manager) error {
	return manager.EnsureDependency(ctx, NewHomebrewDependency())
}

// InstallFormula installs a Homebrew formula
func (i *DarwinInstaller) InstallFormula(ctx context.Context, formula string) error {
	if !HasBrew() {
		return fmt.Errorf("homebrew is not installed")
	}

	output, err := RunCommand(ctx, "brew", "install", formula)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", formula, err, output)
	}

	return nil
}

// InstallCask installs a Homebrew cask
func (i *DarwinInstaller) InstallCask(ctx context.Context, cask string) error {
	if !HasBrew() {
		return fmt.Errorf("homebrew is not installed")
	}

	output, err := RunCommand(ctx, "brew", "install", "--cask", cask)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", cask, err, output)
	}

	return nil
}

// PrintHomebrewPathInstructions prints instructions for adding Homebrew to PATH
func PrintHomebrewPathInstructions() {
	platform := DetectPlatform()

	fmt.Fprintln(os.Stderr, "\n=== Homebrew PATH Configuration ===")
	fmt.Fprintln(os.Stderr, "Add the following to your shell profile (~/.zshrc or ~/.bash_profile):")
	fmt.Fprintln(os.Stderr)

	if platform.IsAppleSilicon() {
		fmt.Fprintln(os.Stderr, `  eval "$(/opt/homebrew/bin/brew shellenv)"`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Homebrew is installed at: /opt/homebrew")
	} else {
		fmt.Fprintln(os.Stderr, `  eval "$(/usr/local/bin/brew shellenv)"`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Homebrew is installed at: /usr/local")
	}

	fmt.Fprintln(os.Stderr, "\nThen restart your terminal or run the command above.")
	fmt.Fprintln(os.Stderr, "===================================")
}
