package deps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	// containerReleaseAPI is the GitHub API endpoint listing the latest container release.
	containerReleaseAPI = "https://api.github.com/repos/apple/container/releases/latest"

	// containerReleasesPage is where a manual install can be done.
	containerReleasesPage = "https://github.com/apple/container/releases"

	// containerInstallerAsset is the asset name suffix of the signed installer package.
	containerInstallerAsset = "-installer-signed.pkg"
)

// ContainerDependency represents Apple's "container" runtime.
type ContainerDependency struct{}

// NewContainerDependency creates a new container dependency
func NewContainerDependency() *ContainerDependency {
	return &ContainerDependency{}
}

// Name returns the dependency name
func (c *ContainerDependency) Name() string {
	return "Apple container"
}

// Check returns the status of the container runtime: not installed, or installed
// with its services running or stopped.
func (c *ContainerDependency) Check(ctx context.Context) (Status, error) {
	if !CommandExists("container") {
		return StatusNotInstalled, nil
	}

	// "container system status" exits non-zero when the services are not running.
	if err := exec.CommandContext(ctx, "container", "system", "status").Run(); err != nil {
		return StatusStopped, nil
	}

	return StatusRunning, nil
}

// Install downloads the signed installer package and installs it with the macOS installer
func (c *ContainerDependency) Install(ctx context.Context) error {
	if !DetectPlatform().IsDarwin() {
		return fmt.Errorf("container installation is only supported on macOS")
	}

	url, err := latestInstallerURL(ctx)
	if err != nil {
		return err
	}

	pkgPath, err := downloadInstaller(ctx, url)
	if err != nil {
		return err
	}
	defer os.Remove(pkgPath)

	// The package installs into /usr/local, which requires administrator rights.
	cmd := exec.CommandContext(ctx, "sudo", "installer", "-pkg", pkgPath, "-target", "/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install container: %w", err)
	}

	return nil
}

// Start starts the container services and installs the recommended kernel if missing
func (c *ContainerDependency) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "container", "system", "start", "--enable-kernel-install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container services: %w", err)
	}

	return nil
}

// RequiresStart returns true as the container services must be running
func (c *ContainerDependency) RequiresStart() bool {
	return true
}

// GetVersion returns the installed container version
func (c *ContainerDependency) GetVersion(ctx context.Context) (string, error) {
	output, err := GetCommandVersion(ctx, "container", "--version")
	if err != nil {
		return "", err
	}

	// Output format: "container CLI version 1.6.0 (build: release, commit: ...)"
	re := regexp.MustCompile(`version (\S+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	return strings.TrimSpace(output), nil
}

// latestInstallerURL resolves the download URL of the latest signed installer package
func latestInstallerURL(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, containerReleaseAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query container releases: %w (install manually from %s)", err, containerReleasesPage)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to query container releases: %s (install manually from %s)", resp.Status, containerReleasesPage)
	}

	var release struct {
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse container releases: %w (install manually from %s)", err, containerReleasesPage)
	}

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, containerInstallerAsset) {
			return asset.URL, nil
		}
	}

	return "", fmt.Errorf("no installer package found (install manually from %s)", containerReleasesPage)
}

// downloadInstaller downloads the installer package into a temporary file
func downloadInstaller(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download container installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download container installer: %s", resp.Status)
	}

	file, err := os.CreateTemp("", "container-*.pkg")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(file.Name())
		return "", fmt.Errorf("failed to save container installer: %w", err)
	}

	return file.Name(), nil
}
