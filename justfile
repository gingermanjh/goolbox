# Build and quality tasks for the ffmpeg-converter CLI.
#
# Run `just` to list the available recipes.

set shell := ["bash", "-o", "pipefail", "-euc"]

app := "cmd/ffmpeg-converter"
binary := "ffmpeg-converter"

version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
commit := `git rev-parse --short HEAD 2>/dev/null || echo unknown`
build_time := `date -u '+%Y-%m-%dT%H:%M:%SZ'`
ldflags := "-s -w -X main.Version=" + version + " -X main.Commit=" + commit + " -X main.BuildTime=" + build_time

# staticcheck is built with an older Go and cannot read the export data of a
# newer toolchain, so it runs on the version declared in go.mod.
staticcheck_toolchain := "go" + `awk '/^go /{print $2}' go.mod`

# List the available recipes.
default:
    @just --list

# Run all tests with race detection.
test:
    go test -race ./...

# Run the tests verbosely.
test-verbose:
    go test -race -v ./...

# Run the tests and write a coverage report.
cover:
    go test -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Format the code and tidy go.mod.
tidy:
    go mod tidy
    go fmt ./...

# Run go vet.
vet:
    go vet ./...

# Run staticcheck.
lint:
    GOTOOLCHAIN={{staticcheck_toolchain}} $(go env GOPATH)/bin/staticcheck ./...

# Run every quality check.
audit: tidy vet lint test

# Build the ffmpeg-converter binary.
build:
    cd {{app}} && go build -trimpath -ldflags "{{ldflags}}" -o {{binary}} .

# Build the binary and show the CLI help.
run: build
    cd {{app}} && ./{{binary}} -h

# Install the binary into GOPATH/bin.
install:
    cd {{app}} && go install -trimpath -ldflags "{{ldflags}}" .

# Print the version information embedded in the binary.
version:
    @echo "Version:    {{version}}"
    @echo "Commit:     {{commit}}"
    @echo "Build Time: {{build_time}}"

# Remove the build artifacts and the coverage files.
clean:
    cd {{app}} && rm -f {{binary}} && rm -rf build
    rm -f coverage.out coverage.html
