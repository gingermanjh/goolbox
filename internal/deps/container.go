package deps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	// containerUserInstallDir is the directory, relative to the home directory,
	// where the runtime is installed when no system wide install is present.
	containerUserInstallDir = ".local/container"

	// containerFallbackVersion is installed when the release API is unavailable.
	// The API is rate limited per IP, which a shared network can hit.
	containerFallbackVersion = "1.6.0"
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
	path := ContainerPath()
	if path == "" {
		return StatusNotInstalled, nil
	}

	// "container system status" exits non-zero when the services are not running.
	if err := exec.CommandContext(ctx, path, "system", "status").Run(); err != nil {
		return StatusStopped, nil
	}

	return StatusRunning, nil
}

// Install installs the runtime into the home directory. The official signed
// package is downloaded and unpacked, which keeps the install free of
// administrator rights.
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

	installDir, err := userInstallDir()
	if err != nil {
		return err
	}

	// Unpack the package instead of running the system wide installer, which
	// would ask for an administrator password.
	extractDir, err := os.MkdirTemp("", "container-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	// pkgutil refuses to expand into an existing directory.
	expandDir := filepath.Join(extractDir, "unpacked")
	if output, err := RunCommand(ctx, "pkgutil", "--expand-full", pkgPath, expandDir); err != nil {
		return fmt.Errorf("failed to unpack container installer: %w (%s)", err, output)
	}

	payload := filepath.Join(expandDir, "Payload")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return err
	}

	// The CLI looks for its plugins in <install root>/libexec/container/plugins,
	// so the payload has to keep the layout of the package.
	for _, dir := range []string{"bin", "libexec"} {
		source := filepath.Join(payload, dir)
		if _, err := os.Stat(source); err != nil {
			return fmt.Errorf("container installer is missing %s: %w", dir, err)
		}
		if output, err := RunCommand(ctx, "cp", "-R", source, filepath.Join(installDir, dir)); err != nil {
			return fmt.Errorf("failed to install container: %w (%s)", err, output)
		}
	}

	return nil
}

// Start starts the container services and installs the recommended kernel if missing
func (c *ContainerDependency) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, ContainerPath(), "system", "start", "--enable-kernel-install")
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

// Stop stops the container services so no runtime resources stay behind
func (c *ContainerDependency) Stop(ctx context.Context) error {
	output, err := RunCommand(ctx, ContainerPath(), "system", "stop")
	if err != nil {
		return fmt.Errorf("failed to stop container services: %w (%s)", err, output)
	}

	return nil
}

// PruneStaleImages removes untagged images left behind by older versions of a
// tag. The image in use is kept so the next run does not have to pull it again.
// The container services must be running for this to work.
func (c *ContainerDependency) PruneStaleImages(ctx context.Context) error {
	output, err := RunCommand(ctx, ContainerPath(), "image", "prune")
	if err != nil {
		return fmt.Errorf("failed to prune stale container images: %w (%s)", err, output)
	}

	return nil
}

// GetVersion returns the installed container version
func (c *ContainerDependency) GetVersion(ctx context.Context) (string, error) {
	output, err := GetCommandVersion(ctx, ContainerPath(), "--version")
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

// userInstallDir returns the directory the runtime is installed in when there
// is no system wide install.
func userInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to locate the home directory: %w", err)
	}
	return filepath.Join(home, containerUserInstallDir), nil
}

// containerPath returns the path of the container CLI. A system wide install
// wins over the user local one, and an empty string means that neither is
// present.
func ContainerPath() string {
	if path, err := exec.LookPath("container"); err == nil {
		return path
	}

	dir, err := userInstallDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(dir, "bin", "container")
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

// latestInstallerURL resolves the download URL of the signed installer package.
// The release API gives the newest version, but it is rate limited per IP, so
// a known version is used as the fallback.
func latestInstallerURL(ctx context.Context) (string, error) {
	url, err := releaseInstallerURL(ctx)
	if err != nil {
		return containerInstallerURL(containerFallbackVersion), nil
	}
	return url, nil
}

// releaseInstallerURL asks the release API for the newest signed installer.
func releaseInstallerURL(ctx context.Context) (string, error) {
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

// containerInstallerURL builds the download URL of a signed installer package.
func containerInstallerURL(version string) string {
	return fmt.Sprintf("https://github.com/apple/container/releases/download/%s/container-%s%s",
		version, version, containerInstallerAsset)
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
