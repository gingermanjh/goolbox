# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project with multiple CLI applications. Each app has its own Makefile for building, while the root Makefile handles project-wide tasks (testing, linting, etc.).

### Applications

| App | Description | Location |
|-----|-------------|----------|
| ffmpeg-converter | Video conversion tool using ffmpeg (Docker or local) | `cmd/ffmpeg-converter/` |

## Project Structure

```
goolbox/
├── cmd/                         # Application entry points
│   └── ffmpeg-converter/        # Video converter CLI
│       ├── main.go
│       ├── Makefile             # App-specific build
│       ├── README.md
│       └── USAGE.md
├── internal/                    # Internal packages
│   ├── ffmpeg/                  # ffmpeg conversion logic
│   └── deps/                    # Dependency auto-management
├── Makefile                     # Project-wide build/test
├── CLAUDE.md
└── go.mod
```

## Build Commands

### Using Makefiles (Recommended)

```bash
# Build all apps (from root)
make build

# Build specific app
make -C cmd/ffmpeg-converter build
# or
cd cmd/ffmpeg-converter && make build

# Cross-compile for all platforms
make build/all
```

### Using Go directly

```bash
# Run app
go run ./cmd/ffmpeg-converter -h

# Build app
go build -o ffmpeg-converter ./cmd/ffmpeg-converter
```

## Testing and Quality

```bash
# Run all tests
make test

# Run tests with coverage
make test/cover

# Run linting (staticcheck)
make lint

# Run go vet
make vet

# Run all quality checks (tidy + vet + lint + test)
make audit

# Format and tidy
make tidy
```

## Adding a New App

1. Create directory: `cmd/<app-name>/`
2. Add `main.go` and copy `Makefile` from existing app
3. Update `BINARY_NAME` in the new Makefile
4. Add app name to `APPS` variable in root `Makefile`

## Code Style

- All comments in `.go` files must be written in English
- Follow standard Go conventions (gofmt, go vet)
- Use meaningful variable and function names
