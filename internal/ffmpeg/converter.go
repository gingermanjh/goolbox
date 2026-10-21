package ffmpeg

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ContainerImage is the container image that provides ffmpeg.
const ContainerImage = "linuxserver/ffmpeg:latest"

type Config struct {
	// ContainerPath is the path of the container CLI. It is empty when the
	// runtime is only reachable through PATH.
	ContainerPath string
	CRF           string
	Preset        string
	Scale         string
	IncludeAudio  bool
	DryRun        bool
	// Speed is the playback speed multiplier. 1.0 = normal, 2.0 = twice as fast,
	// 0.5 = half speed. Values <= 0 are treated as 1.0 (no change).
	Speed float64
}

type Converter struct {
	Config Config
}

func NewConverter(cfg Config) *Converter {
	return &Converter{Config: cfg}
}

func (c *Converter) ConvertFile(inputPath, outputPath string) error {
	cmd := c.buildContainerCommand(inputPath, outputPath)

	fmt.Printf("Executing: %s\n", strings.Join(cmd.Args, " "))

	if c.Config.DryRun {
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Converter) buildFFmpegArgs(inputFile, outputFile string) []string {
	// libx264 rate control is driven by CRF alone. Passing -b:v alongside -crf
	// has no effect, so target quality is expressed only through CRF.
	args := []string{
		"-i", inputFile,
		"-c:v", "libx264",
		"-crf", c.Config.CRF,
		"-preset", c.Config.Preset,
		"-vf", c.buildVideoFilter(),
	}

	if c.Config.IncludeAudio {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
		if af := c.buildAudioFilter(); af != "" {
			args = append(args, "-af", af)
		}
	} else {
		args = append(args, "-an")
	}

	args = append(args, "-y", outputFile)
	return args
}

func (c *Converter) buildVideoFilter() string {
	// The profile gives a target size, not the output size: the picture keeps
	// its shape and fits inside the box instead of being stretched.
	scale := fmt.Sprintf("scale=%s:force_original_aspect_ratio=decrease", c.Config.Scale)
	if !c.hasSpeedChange() {
		return scale
	}
	// setpts multiplier is the inverse of the playback speed.
	return fmt.Sprintf("setpts=%s*PTS,%s", formatFloat(1.0/c.Config.Speed), scale)
}

func (c *Converter) buildAudioFilter() string {
	if !c.hasSpeedChange() {
		return ""
	}
	return buildAtempoChain(c.Config.Speed)
}

func (c *Converter) hasSpeedChange() bool {
	return c.Config.Speed > 0 && math.Abs(c.Config.Speed-1.0) > 1e-9
}

// buildAtempoChain produces an atempo filter chain that achieves the requested
// playback speed. A single atempo filter only accepts factors in [0.5, 100.0];
// to support smaller or larger factors we chain multiple instances.
func buildAtempoChain(speed float64) string {
	var filters []string
	remaining := speed
	for remaining > 2.0 {
		filters = append(filters, "atempo=2.0")
		remaining /= 2.0
	}
	for remaining < 0.5 {
		filters = append(filters, "atempo=0.5")
		remaining /= 0.5
	}
	if math.Abs(remaining-1.0) > 1e-9 {
		filters = append(filters, "atempo="+formatFloat(remaining))
	}
	return strings.Join(filters, ",")
}

func formatFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", v), "0"), ".")
}

// buildContainerCommand builds the Apple container command that runs ffmpeg
// inside the container image with the input and output directories mounted.
func (c *Converter) buildContainerCommand(inputPath, outputPath string) *exec.Cmd {
	inputDir := filepath.Dir(inputPath)
	outputDir := filepath.Dir(outputPath)
	inputFile := filepath.Base(inputPath)
	outputFile := filepath.Base(outputPath)

	ffmpegArgs := c.buildFFmpegArgs("/input/"+inputFile, "/output/"+outputFile)

	containerArgs := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/input", inputDir),
		"-v", fmt.Sprintf("%s:/output", outputDir),
		ContainerImage,
	}
	containerArgs = append(containerArgs, ffmpegArgs...)

	binary := c.Config.ContainerPath
	if binary == "" {
		binary = "container"
	}

	return exec.Command(binary, containerArgs...)
}

func FindFiles(dir, ext string) ([]string, error) {
	pattern := filepath.Join(dir, "*."+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	return matches, nil
}

func GenerateOutputPath(inputPath, outputDir string) string {
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	return filepath.Join(outputDir, nameWithoutExt+".mp4")
}
