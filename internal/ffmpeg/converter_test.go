package ffmpeg

import (
	"reflect"
	"testing"
)

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{1.0, "1"},
		{2.0, "2"},
		{0.5, "0.5"},
		{1.25, "1.25"},
		{0.666667, "0.666667"},
		{4.0, "4"},
	}
	for _, tc := range tests {
		if got := formatFloat(tc.in); got != tc.want {
			t.Errorf("formatFloat(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildAtempoChain(t *testing.T) {
	tests := []struct {
		name  string
		speed float64
		want  string
	}{
		{"within range 1.5x", 1.5, "atempo=1.5"},
		{"within range 0.75x", 0.75, "atempo=0.75"},
		{"boundary 2x stays as-is", 2.0, "atempo=2"},
		{"boundary 0.5x stays as-is", 0.5, "atempo=0.5"},
		{"5x chains over upper bound", 5.0, "atempo=2.0,atempo=2.0,atempo=1.25"},
		{"4x chains over upper bound", 4.0, "atempo=2.0,atempo=2"},
		{"0.25x chains under lower bound", 0.25, "atempo=0.5,atempo=0.5"},
		{"1x produces no filter", 1.0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildAtempoChain(tc.speed)
			if got != tc.want {
				t.Errorf("buildAtempoChain(%v) = %q, want %q", tc.speed, got, tc.want)
			}
		})
	}
}

func TestHasSpeedChange(t *testing.T) {
	tests := []struct {
		name  string
		speed float64
		want  bool
	}{
		{"zero treated as no change", 0, false},
		{"negative treated as no change", -1, false},
		{"exactly 1.0", 1.0, false},
		{"within float tolerance", 1.0 + 1e-12, false},
		{"2x", 2.0, true},
		{"0.5x", 0.5, true},
		{"1.5x", 1.5, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Converter{Config: Config{Speed: tc.speed}}
			if got := c.hasSpeedChange(); got != tc.want {
				t.Errorf("hasSpeedChange(speed=%v) = %v, want %v", tc.speed, got, tc.want)
			}
		})
	}
}

func TestBuildVideoFilter(t *testing.T) {
	tests := []struct {
		name  string
		speed float64
		scale string
		want  string
	}{
		{"no speed change keeps only scale", 1.0, "1920:1080", "scale=1920:1080"},
		{"zero speed treated as no change", 0, "1920:1080", "scale=1920:1080"},
		{"2x prepends setpts", 2.0, "1920:1080", "setpts=0.5*PTS,scale=1920:1080"},
		{"0.5x prepends setpts", 0.5, "1280:720", "setpts=2*PTS,scale=1280:720"},
		{"1.5x prepends setpts with rounded multiplier", 1.5, "1920:1080", "setpts=0.666667*PTS,scale=1920:1080"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Converter{Config: Config{Speed: tc.speed, Scale: tc.scale}}
			if got := c.buildVideoFilter(); got != tc.want {
				t.Errorf("buildVideoFilter() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildAudioFilter(t *testing.T) {
	tests := []struct {
		name  string
		speed float64
		want  string
	}{
		{"no speed change returns empty", 1.0, ""},
		{"zero speed returns empty", 0, ""},
		{"2x", 2.0, "atempo=2"},
		{"1.5x", 1.5, "atempo=1.5"},
		{"5x chains", 5.0, "atempo=2.0,atempo=2.0,atempo=1.25"},
		{"0.25x chains", 0.25, "atempo=0.5,atempo=0.5"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Converter{Config: Config{Speed: tc.speed}}
			if got := c.buildAudioFilter(); got != tc.want {
				t.Errorf("buildAudioFilter() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildFFmpegArgs(t *testing.T) {
	baseConfig := Config{
		CRF:     "18",
		Preset:  "slow",
		Bitrate: "3000k",
		Scale:   "1920:1080",
	}

	tests := []struct {
		name   string
		config Config
		want   []string
	}{
		{
			name: "standard speed without audio",
			config: func() Config {
				c := baseConfig
				c.Speed = 1.0
				return c
			}(),
			want: []string{
				"-i", "in.avi",
				"-c:v", "libx264",
				"-crf", "18",
				"-preset", "slow",
				"-b:v", "3000k",
				"-vf", "scale=1920:1080",
				"-an",
				"-y", "out.mp4",
			},
		},
		{
			name: "2x speed without audio",
			config: func() Config {
				c := baseConfig
				c.Speed = 2.0
				return c
			}(),
			want: []string{
				"-i", "in.avi",
				"-c:v", "libx264",
				"-crf", "18",
				"-preset", "slow",
				"-b:v", "3000k",
				"-vf", "setpts=0.5*PTS,scale=1920:1080",
				"-an",
				"-y", "out.mp4",
			},
		},
		{
			name: "1.5x speed with audio",
			config: func() Config {
				c := baseConfig
				c.Speed = 1.5
				c.IncludeAudio = true
				return c
			}(),
			want: []string{
				"-i", "in.avi",
				"-c:v", "libx264",
				"-crf", "18",
				"-preset", "slow",
				"-b:v", "3000k",
				"-vf", "setpts=0.666667*PTS,scale=1920:1080",
				"-c:a", "aac", "-b:a", "128k",
				"-af", "atempo=1.5",
				"-y", "out.mp4",
			},
		},
		{
			name: "5x speed with audio uses atempo chain",
			config: func() Config {
				c := baseConfig
				c.Speed = 5.0
				c.IncludeAudio = true
				return c
			}(),
			want: []string{
				"-i", "in.avi",
				"-c:v", "libx264",
				"-crf", "18",
				"-preset", "slow",
				"-b:v", "3000k",
				"-vf", "setpts=0.2*PTS,scale=1920:1080",
				"-c:a", "aac", "-b:a", "128k",
				"-af", "atempo=2.0,atempo=2.0,atempo=1.25",
				"-y", "out.mp4",
			},
		},
		{
			name: "standard speed with audio omits -af",
			config: func() Config {
				c := baseConfig
				c.Speed = 1.0
				c.IncludeAudio = true
				return c
			}(),
			want: []string{
				"-i", "in.avi",
				"-c:v", "libx264",
				"-crf", "18",
				"-preset", "slow",
				"-b:v", "3000k",
				"-vf", "scale=1920:1080",
				"-c:a", "aac", "-b:a", "128k",
				"-y", "out.mp4",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Converter{Config: tc.config}
			got := c.buildFFmpegArgs("in.avi", "out.mp4")
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("buildFFmpegArgs() mismatch\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestGenerateOutputPath(t *testing.T) {
	tests := []struct {
		input     string
		outputDir string
		want      string
	}{
		{"/tmp/video.avi", "/out", "/out/video.mp4"},
		{"/tmp/video.mkv", "/out", "/out/video.mp4"},
		{"/path/to/my.video.avi", "/out", "/out/my.video.mp4"},
	}
	for _, tc := range tests {
		if got := GenerateOutputPath(tc.input, tc.outputDir); got != tc.want {
			t.Errorf("GenerateOutputPath(%q, %q) = %q, want %q", tc.input, tc.outputDir, got, tc.want)
		}
	}
}
