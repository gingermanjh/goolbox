package ffmpeg

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	UseDocker    bool
	CRF          string
	Preset       string
	Bitrate      string
	Scale        string
	IncludeAudio bool
	DryRun       bool
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
	var cmd *exec.Cmd

	if c.Config.UseDocker {
		cmd = c.buildDockerCommand(inputPath, outputPath)
	} else {
		cmd = c.buildLocalCommand(inputPath, outputPath)
	}

	fmt.Printf("Executing: %s\n", strings.Join(cmd.Args, " "))

	if c.Config.DryRun {
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Converter) buildFFmpegArgs(inputFile, outputFile string) []string {
	args := []string{
		"-i", inputFile,
		"-c:v", "libx264",
		"-crf", c.Config.CRF,
		"-preset", c.Config.Preset,
		"-b:v", c.Config.Bitrate,
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
	scale := fmt.Sprintf("scale=%s", c.Config.Scale)
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

func (c *Converter) buildDockerCommand(inputPath, outputPath string) *exec.Cmd {
	inputDir := filepath.Dir(inputPath)
	outputDir := filepath.Dir(outputPath)
	inputFile := filepath.Base(inputPath)
	outputFile := filepath.Base(outputPath)

	ffmpegArgs := c.buildFFmpegArgs("/input/"+inputFile, "/output/"+outputFile)

	dockerArgs := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/input", inputDir),
		"-v", fmt.Sprintf("%s:/output", outputDir),
		"linuxserver/ffmpeg:latest",
	}
	dockerArgs = append(dockerArgs, ffmpegArgs...)

	return exec.Command("docker", dockerArgs...)
}

func (c *Converter) buildLocalCommand(inputPath, outputPath string) *exec.Cmd {
	ffmpegArgs := c.buildFFmpegArgs(inputPath, outputPath)
	return exec.Command("ffmpeg", ffmpegArgs...)
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
