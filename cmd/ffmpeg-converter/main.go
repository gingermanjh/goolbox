package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	os.Exit(run())
}

func run() int {
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
	includeAudio := flag.Bool("audio", false, "Include audio in output (overrides profile default)")
	speed := flag.Float64("speed", 0, "Playback speed multiplier (e.g. 2.0 = 2x, 0.5 = half). 0 = use profile default")
	dryRun := flag.Bool("dry-run", false, "Print commands without executing")

	// Profile flags
	profileName := flag.String("profile", ffmpeg.DefaultProfile, "Encoding profile to use")
	listProfiles := flag.Bool("list-profiles", false, "List available encoding profiles")

	// Dependency management flags
	autoInstall := flag.Bool("auto-install", true, "Automatically install missing dependencies (default: enabled)")

	// Cleanup flags
	keepServices := flag.Bool("keep-services", false, "Keep the container services running after the conversion")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Convert video files using ffmpeg in an Apple container\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input ./videos -output ./results\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input ./videos -ext mkv -audio\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -profile 2x -input ./videos\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -speed 1.5 -input ./videos\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-profiles\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input ./videos -keep-services\n", os.Args[0])
	}

	flag.Parse()

	// Handle list-profiles flag
	if *listProfiles {
		printProfiles()
		return 0
	}

	// Handle version flag
	if *showVersion {
		fmt.Printf("ffmpeg-converter %s (commit: %s, built: %s)\n", Version, Commit, BuildTime)
		return 0
	}

	// Get the selected profile
	profile, ok := ffmpeg.GetProfile(*profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown profile: %s\n", *profileName)
		fmt.Fprintf(os.Stderr, "Use -list-profiles to see available profiles\n")
		return 1
	}

	// Apple container only runs on macOS 26 or later on Apple silicon.
	if !deps.SupportsContainer() {
		fmt.Fprintf(os.Stderr, "This tool requires macOS %d or later on Apple silicon.\n", deps.SupportedContainerMacOSMajor)
		return 1
	}

	container := deps.NewContainerDependency()

	// Check and install dependencies if needed
	if err := ensureDependencies(container, *autoInstall); err != nil {
		fmt.Fprintf(os.Stderr, "Dependency error: %v\n", err)
		return 1
	}

	// Release the container runtime resources when the conversion is done.
	defer releaseResources(container, *keepServices)

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
		return 1
	}

	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return 1
	}

	files, err := ffmpeg.FindFiles(absInputDir, *ext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		return 1
	}

	if len(files) == 0 {
		fmt.Printf("No .%s files found in %s\n", *ext, absInputDir)
		return 0
	}

	fmt.Printf("Found %d .%s file(s) to convert\n", len(files), *ext)
	fmt.Printf("Input directory: %s\n", absInputDir)
	fmt.Printf("Output directory: %s\n", absOutputDir)
	fmt.Println("Mode: Apple container")
	fmt.Println()

	// Determine audio setting: CLI flag overrides profile default
	audioEnabled := profile.Audio
	if *includeAudio {
		audioEnabled = true
	}

	// Determine speed: CLI flag (when > 0) overrides profile default
	if *speed < 0 {
		fmt.Fprintf(os.Stderr, "Speed must not be negative (0 means the profile value).\n")
		return 1
	}
	speedValue := profile.Speed
	if *speed > 0 {
		speedValue = *speed
	}

	converter := ffmpeg.NewConverter(ffmpeg.Config{
		ContainerPath: deps.ContainerPath(),
		CRF:           profile.CRF,
		Preset:        profile.Preset,
		Scale:         profile.Scale,
		IncludeAudio:  audioEnabled,
		DryRun:        *dryRun,
		Speed:         speedValue,
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
		return 1
	}
	return 0
}

// releaseResources removes the container runtime resources left behind by the
// conversion: stale images and the background services.
func releaseResources(container *deps.ContainerDependency, keepServices bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Only untagged images are removed, so the image in use stays cached for
	// the next run. Pruning talks to the container services, so it runs first.
	if err := container.PruneStaleImages(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if keepServices {
		return
	}

	if err := container.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	} else {
		fmt.Println("Container services stopped")
	}
}

// ensureDependencies checks and installs the Apple container runtime
func ensureDependencies(container *deps.ContainerDependency, autoInstall bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	manager := deps.NewManager(autoInstall, os.Stdout)

	if err := manager.EnsureDependency(ctx, container); err != nil {
		return fmt.Errorf("container: %w", err)
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
