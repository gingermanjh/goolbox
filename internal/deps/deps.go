package deps

import (
	"context"
	"fmt"
	"io"
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
	output      io.Writer
}

// NewManager creates a new dependency manager
func NewManager(autoInstall bool, output io.Writer) *Manager {
	return &Manager{
		autoInstall: autoInstall,
		output:      output,
	}
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
	// Starting an installed dependency is part of the normal run, so it needs
	// neither a confirmation nor the install to be allowed.
	fmt.Fprintf(m.output, "Starting %s...\n", dep.Name())
	if err := dep.Start(ctx); err != nil {
		return fmt.Errorf("failed to start %s: %w", dep.Name(), err)
	}

	fmt.Fprintf(m.output, "✓ %s started successfully\n", dep.Name())
	return nil
}
