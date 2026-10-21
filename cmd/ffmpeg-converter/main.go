package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"goolbox/internal/deps"
	"goolbox/internal/ffmpeg"
)

// Version information (set via ldflags at build time)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Version flag (check early before other processing)
	showVersion := flag.Bool("version", false, "Show version information")

	// Use executable location as the default path (so it works when double-clicked in Finder)
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)

	inputDir := flag.String("input", execDir, "Input directory containing video files")
	outputDir := flag.String("output", filepath.Join(execDir, "results"), "Output directory for converted files")
	ext := flag.String("ext", "avi", "Input file extension to search for")
	useDocker := flag.Bool("docker", true, "Use Docker ffmpeg (default)")
	useLocal := flag.Bool("local", false, "Use local ffmpeg instead of Docker")
	includeAudio := flag.Bool("audio", false, "Include audio in output (overrides profile default)")
	speed := flag.Float64("speed", 0, "Playback speed multiplier (e.g. 2.0 = 2x, 0.5 = half). 0 = use profile default")
	dryRun := flag.Bool("dry-run", false, "Print commands without executing")

	// Profile flags
	profileName := flag.String("profile", ffmpeg.DefaultProfile, "Encoding profile to use")
	listProfiles := flag.Bool("list-profiles", false, "List available encoding profiles")

	// Dependency management flags
	autoInstall := flag.Bool("auto-install", true, "Automatically install missing dependencies (default: enabled)")
	noAutoInstall := flag.Bool("no-auto-install", false, "Skip dependency checking entirely")
	skipConfirm := flag.Bool("y", false, "Skip confirmation prompts for installation")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Convert video files using ffmpeg (via Docker or local installation)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input ./videos -output ./results\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input ./videos -ext mkv -local -audio\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -profile 2x -input ./videos\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -speed 1.5 -input ./videos\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-profiles\n", os.Args[0])
	}

	flag.Parse()

	// Handle list-profiles flag
	if *listProfiles {
		printProfiles()
		os.Exit(0)
	}

	// Handle version flag
	if *showVersion {
		fmt.Printf("ffmpeg-converter %s (commit: %s, built: %s)\n", Version, Commit, BuildTime)
		os.Exit(0)
	}

	// Get the selected profile
	profile, ok := ffmpeg.GetProfile(*profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown profile: %s\n", *profileName)
		fmt.Fprintf(os.Stderr, "Use -list-profiles to see available profiles\n")
		os.Exit(1)
	}

	dockerMode := *useDocker && !*useLocal

	// Check and install dependencies if needed
	if !*noAutoInstall {
		if err := ensureDependencies(dockerMode, *autoInstall, *skipConfirm); err != nil {
			fmt.Fprintf(os.Stderr, "Dependency error: %v\n", err)
			os.Exit(1)
		}
	}

	absInputDir, err := filepath.Abs(*inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving input path: %v\n", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving output path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absInputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Input directory does not exist: %s\n", absInputDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	files, err := ffmpeg.FindFiles(absInputDir, *ext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No .%s files found in %s\n", *ext, absInputDir)
		os.Exit(0)
	}

	fmt.Printf("Found %d .%s file(s) to convert\n", len(files), *ext)
	fmt.Printf("Input directory: %s\n", absInputDir)
	fmt.Printf("Output directory: %s\n", absOutputDir)
	if dockerMode {
		fmt.Println("Mode: Docker")
	} else {
		fmt.Println("Mode: Local ffmpeg")
	}
	fmt.Println()

	// Determine audio setting: CLI flag overrides profile default
	audioEnabled := profile.Audio
	if *includeAudio {
		audioEnabled = true
	}

	// Determine speed: CLI flag (when > 0) overrides profile default
	speedValue := profile.Speed
	if *speed > 0 {
		speedValue = *speed
	}

	converter := ffmpeg.NewConverter(ffmpeg.Config{
		UseDocker:    dockerMode,
		CRF:          profile.CRF,
		Preset:       profile.Preset,
		Scale:        profile.Scale,
		IncludeAudio: audioEnabled,
		DryRun:       *dryRun,
		Speed:        speedValue,
	})

	successCount := 0
	failCount := 0

	for i, inputFile := range files {
		outputFile := ffmpeg.GenerateOutputPath(inputFile, absOutputDir)
		fmt.Printf("[%d/%d] Converting: %s\n", i+1, len(files), filepath.Base(inputFile))

		if err := converter.ConvertFile(inputFile, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error converting %s: %v\n", inputFile, err)
			failCount++
			continue
		}

		fmt.Printf("Successfully created: %s\n\n", outputFile)
		successCount++
	}

	fmt.Println("========================================")
	fmt.Printf("Conversion complete: %d succeeded, %d failed\n", successCount, failCount)

	if failCount > 0 {
		os.Exit(1)
	}
}

// ensureDependencies checks and installs required dependencies based on the mode
func ensureDependencies(dockerMode, autoInstall, skipConfirm bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	manager := deps.NewManager(autoInstall, skipConfirm, os.Stdout)

	// Set up input for confirmation prompts
	if !skipConfirm {
		manager.SetInput(bufio.NewReader(os.Stdin))
	}

	// On macOS, ensure Homebrew is available first (needed for installing other deps)
	if runtime.GOOS == "darwin" {
		if err := manager.EnsureDependency(ctx, deps.NewHomebrewDependency()); err != nil {
			return fmt.Errorf("homebrew: %w", err)
		}
	}

	// Check/install the appropriate dependency based on mode
	if dockerMode {
		if err := manager.EnsureDependency(ctx, deps.NewDockerDependency()); err != nil {
			return fmt.Errorf("docker: %w", err)
		}
	} else {
		if err := manager.EnsureDependency(ctx, deps.NewFFmpegDependency()); err != nil {
			return fmt.Errorf("ffmpeg: %w", err)
		}
	}

	return nil
}

// printProfiles displays all available encoding profiles
func printProfiles() {
	fmt.Println("Available encoding profiles:")
	fmt.Println()
	for _, p := range ffmpeg.ListProfiles() {
		fmt.Printf("  %s\n", p.Name)
		fmt.Printf("    Description: %s\n", p.Description)
		fmt.Printf("    CRF: %s, Preset: %s, Scale: %s, Audio: %v, Speed: %gx\n",
			p.CRF, p.Preset, p.Scale, p.Audio, p.Speed)
		fmt.Println()
	}
}
