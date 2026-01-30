package ffmpeg

import (
	"fmt"
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
		"-vf", fmt.Sprintf("scale=%s", c.Config.Scale),
	}

	if c.Config.IncludeAudio {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	} else {
		args = append(args, "-an")
	}

	args = append(args, "-y", outputFile)
	return args
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
