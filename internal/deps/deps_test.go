package deps

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

// mockDependency is a mock implementation of Dependency for testing
type mockDependency struct {
	name          string
	status        Status
	requiresStart bool
	installCalled bool
	startCalled   bool
	installErr    error
	startErr      error
}

func (m *mockDependency) Name() string {
	return m.name
}

func (m *mockDependency) Check(ctx context.Context) (Status, error) {
	return m.status, nil
}

func (m *mockDependency) Install(ctx context.Context) error {
	m.installCalled = true
	if m.installErr != nil {
		return m.installErr
	}
	return nil
}

func (m *mockDependency) Start(ctx context.Context) error {
	m.startCalled = true
	if m.startErr != nil {
		return m.startErr
	}
	m.status = StatusRunning
	return nil
}

func (m *mockDependency) RequiresStart() bool {
	return m.requiresStart
}

func TestManager_EnsureDependency_AlreadyRunning(t *testing.T) {
	var output bytes.Buffer
	manager := NewManager(true, &output)

	dep := &mockDependency{
		name:          "test-dep",
		status:        StatusRunning,
		requiresStart: true,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if dep.installCalled {
		t.Error("Install should not be called when already running")
	}
	if dep.startCalled {
		t.Error("Start should not be called when already running")
	}

	if !strings.Contains(output.String(), "is running") {
		t.Errorf("expected 'is running' in output, got: %s", output.String())
	}
}

func TestManager_EnsureDependency_NotInstalled_AutoInstallDisabled(t *testing.T) {
	var output bytes.Buffer
	manager := NewManager(false, &output)

	dep := &mockDependency{
		name:   "test-dep",
		status: StatusNotInstalled,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err == nil {
		t.Error("expected error when auto-install is disabled")
	}

	if !strings.Contains(err.Error(), "-auto-install") {
		t.Errorf("expected error to mention -auto-install flag, got: %v", err)
	}
}

func TestManager_EnsureDependency_NotInstalled_AutoInstallEnabled(t *testing.T) {
	var output bytes.Buffer
	manager := NewManager(true, &output)

	dep := &mockDependency{
		name:   "test-dep",
		status: StatusNotInstalled,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !dep.installCalled {
		t.Error("Install should be called when not installed and auto-install enabled")
	}

	if !strings.Contains(output.String(), "Installing") {
		t.Errorf("expected 'Installing' in output, got: %s", output.String())
	}
}

func TestManager_EnsureDependency_Stopped_AutoInstallEnabled(t *testing.T) {
	var output bytes.Buffer
	manager := NewManager(true, &output)

	dep := &mockDependency{
		name:          "test-dep",
		status:        StatusStopped,
		requiresStart: true,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !dep.startCalled {
		t.Error("Start should be called when stopped")
	}
	if dep.installCalled {
		t.Error("Install should not be called when only stopped")
	}
}

func TestManager_ContextTimeout(t *testing.T) {
	var output bytes.Buffer
	manager := NewManager(true, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give time for context to expire
	time.Sleep(5 * time.Millisecond)

	dep := &mockDependency{
		name:   "test-dep",
		status: StatusRunning,
	}

	// Should still work as we check status synchronously
	err := manager.EnsureDependency(ctx, dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestManager_EnsureDependency_StoppedStartsWhenInstallIsDisabled(t *testing.T) {
	var output bytes.Buffer
	manager := NewManager(false, &output)

	dep := &mockDependency{
		name:          "test-dep",
		status:        StatusStopped,
		requiresStart: true,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !dep.startCalled {
		t.Error("Start should be called when the runtime is installed")
	}
}

func TestManager_EnsureDependency_StoppedStartsWithoutConfirmation(t *testing.T) {
	var output bytes.Buffer
	// No input is configured, so a confirmation prompt would fail the run.
	manager := NewManager(true, &output)

	dep := &mockDependency{
		name:          "test-dep",
		status:        StatusStopped,
		requiresStart: true,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !dep.startCalled {
		t.Error("Start should be called without a confirmation prompt")
	}
	if dep.installCalled {
		t.Error("Install should not be called when only stopped")
	}
}

func TestManager_EnsureDependency_NotInstalledInstallsWithoutConfirmation(t *testing.T) {
	var output bytes.Buffer
	// No input is configured, so a confirmation prompt would fail the run.
	manager := NewManager(true, &output)

	dep := &mockDependency{
		name:   "test-dep",
		status: StatusNotInstalled,
	}

	err := manager.EnsureDependency(context.Background(), dep)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !dep.installCalled {
		t.Error("Install should be called without a confirmation prompt")
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusNotInstalled, "not installed"},
		{StatusRunning, "running"},
		{StatusStopped, "stopped"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("Status.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
