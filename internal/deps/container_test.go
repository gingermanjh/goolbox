package deps

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContainerInstallerURL(t *testing.T) {
	want := "https://github.com/apple/container/releases/download/1.6.0/container-1.6.0-installer-signed.pkg"
	if got := containerInstallerURL("1.6.0"); got != want {
		t.Errorf("containerInstallerURL() = %q, want %q", got, want)
	}
}

func TestLatestInstallerURL(t *testing.T) {
	original := containerReleaseAPI
	defer func() { containerReleaseAPI = original }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"assets":[{"name":"container-9.9.9-installer-signed.pkg","browser_download_url":"https://example.test/container-9.9.9-installer-signed.pkg"}]}`)
	}))
	defer server.Close()

	containerReleaseAPI = server.URL

	got, err := latestInstallerURL(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.test/container-9.9.9-installer-signed.pkg"
	if got != want {
		t.Errorf("latestInstallerURL() = %q, want %q", got, want)
	}
}

func TestLatestInstallerURLFallsBackWhenAPIFails(t *testing.T) {
	original := containerReleaseAPI
	defer func() { containerReleaseAPI = original }()

	// The release API is rate limited per IP, so a failure has to fall back to
	// a known version instead of breaking the install.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer server.Close()

	containerReleaseAPI = server.URL

	got, err := latestInstallerURL(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := containerInstallerURL(containerFallbackVersion)
	if got != want {
		t.Errorf("latestInstallerURL() = %q, want %q", got, want)
	}
}

func TestCopyPayload(t *testing.T) {
	payload := t.TempDir()
	for _, dir := range []string{"bin", "libexec/container/plugins"} {
		if err := os.MkdirAll(filepath.Join(payload, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(payload, "bin", "container"), []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(payload, "libexec", "container", "plugins", "plugin"), []byte("plugin"), 0755); err != nil {
		t.Fatal(err)
	}

	installDir := filepath.Join(t.TempDir(), "container")

	// Copying twice must not nest the files one level deeper, and a file from
	// an earlier install must be gone.
	if err := os.MkdirAll(filepath.Join(installDir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(installDir, "bin", "stale")
	if err := os.WriteFile(stale, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := copyPayload(context.Background(), payload, installDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(installDir, "bin", "container")); err != nil {
		t.Errorf("container binary is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installDir, "libexec", "container", "plugins", "plugin")); err != nil {
		t.Errorf("plugin is missing: %v", err)
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("a file from the earlier install is still there")
	}

	entries, err := os.ReadDir(filepath.Join(installDir, "bin"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "bin") {
			t.Errorf("the copy landed one level too deep: %s", entry.Name())
		}
	}
}
