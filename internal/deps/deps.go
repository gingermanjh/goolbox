package deps

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// Status represents the status of a dependency
type Status int

const (
	StatusNotInstalled Status = iota
	StatusInstalled
	StatusRunning
	StatusStopped
)

func (s Status) String() string {
	switch s {
	case StatusNotInstalled:
		return "not installed"
	case StatusInstalled:
		return "installed"
	case StatusRunning:
		return "running"
	case StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Dependency represents a system dependency that can be checked, installed, and started
type Dependency interface {
	// Name returns the human-readable name of the dependency
	Name() string

	// Check returns the current status of the dependency
	Check(ctx context.Context) (Status, error)

	// Install installs the dependency
	Install(ctx context.Context) error

	// Start starts the dependency (if applicable, e.g., the container services)
	Start(ctx context.Context) error

	// RequiresStart returns true if the dependency needs to be started after installation
	RequiresStart() bool
}

// Manager manages dependency checking and installation
type Manager struct {
	autoInstall bool
	skipConfirm bool
	output      io.Writer
	input       io.Reader
}

// NewManager creates a new dependency manager
func NewManager(autoInstall, skipConfirm bool, output io.Writer) *Manager {
	return &Manager{
		autoInstall: autoInstall,
		skipConfirm: skipConfirm,
		output:      output,
		input:       nil, // Will use stdin if needed
	}
}

// SetInput sets the input reader for user prompts (useful for testing)
func (m *Manager) SetInput(input io.Reader) {
	m.input = input
}

// EnsureDependency checks and ensures a dependency is available
func (m *Manager) EnsureDependency(ctx context.Context, dep Dependency) error {
	status, err := dep.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check %s: %w", dep.Name(), err)
	}

	switch status {
	case StatusRunning:
		fmt.Fprintf(m.output, "✓ %s is running\n", dep.Name())
		return nil

	case StatusInstalled:
		if dep.RequiresStart() {
			fmt.Fprintf(m.output, "✓ %s is installed but not running\n", dep.Name())
			return m.startDependency(ctx, dep)
		}
		fmt.Fprintf(m.output, "✓ %s is installed\n", dep.Name())
		return nil

	case StatusStopped:
		fmt.Fprintf(m.output, "⚠ %s is installed but stopped\n", dep.Name())
		return m.startDependency(ctx, dep)

	case StatusNotInstalled:
		fmt.Fprintf(m.output, "✗ %s is not installed\n", dep.Name())
		return m.installDependency(ctx, dep)
	}

	return nil
}

func (m *Manager) installDependency(ctx context.Context, dep Dependency) error {
	if !m.autoInstall {
		return fmt.Errorf("%s is not installed. Use -auto-install to install it automatically", dep.Name())
	}

	if !m.skipConfirm {
		confirmed, err := m.promptConfirm(fmt.Sprintf("Install %s?", dep.Name()))
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("installation of %s cancelled by user", dep.Name())
		}
	}

	fmt.Fprintf(m.output, "Installing %s...\n", dep.Name())
	if err := dep.Install(ctx); err != nil {
		return fmt.Errorf("failed to install %s: %w", dep.Name(), err)
	}

	fmt.Fprintf(m.output, "✓ %s installed successfully\n", dep.Name())

	if dep.RequiresStart() {
		return m.startDependency(ctx, dep)
	}

	return nil
}

func (m *Manager) startDependency(ctx context.Context, dep Dependency) error {
	if !m.autoInstall {
		return fmt.Errorf("%s is not running. Use -auto-install to start it automatically", dep.Name())
	}

	if !m.skipConfirm {
		confirmed, err := m.promptConfirm(fmt.Sprintf("Start %s?", dep.Name()))
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("start of %s cancelled by user", dep.Name())
		}
	}

	fmt.Fprintf(m.output, "Starting %s...\n", dep.Name())
	if err := dep.Start(ctx); err != nil {
		return fmt.Errorf("failed to start %s: %w", dep.Name(), err)
	}

	fmt.Fprintf(m.output, "✓ %s started successfully\n", dep.Name())
	return nil
}

func (m *Manager) promptConfirm(message string) (bool, error) {
	fmt.Fprintf(m.output, "%s [y/N]: ", message)

	reader := m.input
	if reader == nil {
		reader = bufio.NewReader(strings.NewReader("n\n"))
		// In real usage, this would be os.Stdin
		// For safety, default to "no" if no input is configured
		return false, fmt.Errorf("no input available for confirmation prompt")
	}

	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		response := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return response == "y" || response == "yes", nil
	}

	return false, fmt.Errorf("failed to read user input")
}
