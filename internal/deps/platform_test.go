package deps

import (
	"context"
	"runtime"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	p := DetectPlatform()

	if p.OS != runtime.GOOS {
		t.Errorf("DetectPlatform().OS = %v, want %v", p.OS, runtime.GOOS)
	}

	if p.Arch != runtime.GOARCH {
		t.Errorf("DetectPlatform().Arch = %v, want %v", p.Arch, runtime.GOARCH)
	}
}

func TestPlatform_IsDarwin(t *testing.T) {
	tests := []struct {
		platform Platform
		expected bool
	}{
		{Platform{OS: "darwin", Arch: "amd64"}, true},
		{Platform{OS: "darwin", Arch: "arm64"}, true},
		{Platform{OS: "linux", Arch: "amd64"}, false},
		{Platform{OS: "windows", Arch: "amd64"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.platform.OS, func(t *testing.T) {
			if got := tt.platform.IsDarwin(); got != tt.expected {
				t.Errorf("Platform.IsDarwin() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPlatform_IsAppleSilicon(t *testing.T) {
	tests := []struct {
		platform Platform
		expected bool
	}{
		{Platform{OS: "darwin", Arch: "arm64"}, true},
		{Platform{OS: "darwin", Arch: "amd64"}, false},
		{Platform{OS: "linux", Arch: "arm64"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.platform.OS+"-"+tt.platform.Arch, func(t *testing.T) {
			if got := tt.platform.IsAppleSilicon(); got != tt.expected {
				t.Errorf("Platform.IsAppleSilicon() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCommandExists(t *testing.T) {
	// "ls" should exist on all Unix-like systems
	if runtime.GOOS != "windows" {
		if !CommandExists("ls") {
			t.Error("CommandExists('ls') should return true")
		}
	}

	// Non-existent command
	if CommandExists("nonexistent-command-12345") {
		t.Error("CommandExists('nonexistent-command-12345') should return false")
	}
}

func TestGetCommandVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	ctx := context.Background()

	// Test with a common command
	version, err := GetCommandVersion(ctx, "ls", "--version")
	if err != nil {
		// ls --version might not work on macOS, that's OK
		t.Logf("GetCommandVersion('ls', '--version') returned error: %v", err)
	} else if version == "" {
		t.Log("GetCommandVersion returned empty version (might be expected on some systems)")
	}
}

func TestGetBrewPath(t *testing.T) {
	path := GetBrewPath()

	platform := DetectPlatform()
	if platform.IsAppleSilicon() {
		if path != "/opt/homebrew/bin" {
			t.Errorf("GetBrewPath() = %v, want /opt/homebrew/bin for Apple Silicon", path)
		}
	} else if platform.IsDarwin() {
		if path != "/usr/local/bin" {
			t.Errorf("GetBrewPath() = %v, want /usr/local/bin for Intel Mac", path)
		}
	}
}
