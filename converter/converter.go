package converter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Quality string

const (
	QualityLow      Quality = "low"
	QualityMedium   Quality = "medium"
	QualityHigh     Quality = "high"
	QualityLossless Quality = "lossless"
)

var crfMap = map[Quality]int{
	QualityLow:      28,
	QualityMedium:   23,
	QualityHigh:     18,
	QualityLossless: 0,
}

var supportedFormats = map[string]bool{
	"mp4":  true,
	"mkv":  true,
	"webm": true,
	"avi":  true,
	"mov":  true,
}

var codecMap = map[string]string{
	"h264": "libx264",
	"h265": "libx265",
	"vp9":  "libvpx-vp9",
}

type Options struct {
	Input      string
	Output     string
	Format     string
	Quality    Quality
	Resolution string
	Codec      string
}

type ProgressFunc func(percent float64)

func CheckFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH. Install it:\n  macOS:  brew install ffmpeg\n  Ubuntu: sudo apt install ffmpeg\n  Windows: https://ffmpeg.org/download.html")
	}
	return nil
}

func ValidateOptions(opts *Options) error {
	if _, err := os.Stat(opts.Input); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", opts.Input)
	}

	if opts.Format != "" && !supportedFormats[opts.Format] {
		return fmt.Errorf("unsupported format: %s (supported: mp4, mkv, webm, avi, mov)", opts.Format)
	}

	if opts.Codec != "" {
		if _, ok := codecMap[opts.Codec]; !ok {
			return fmt.Errorf("unsupported codec: %s (supported: h264, h265, vp9)", opts.Codec)
		}
	}

	if opts.Quality != "" {
		if _, ok := crfMap[opts.Quality]; !ok {
			return fmt.Errorf("unsupported quality: %s (supported: low, medium, high, lossless)", opts.Quality)
		}
	}

	if opts.Resolution != "" {
		if !isValidResolution(opts.Resolution) {
			return fmt.Errorf("invalid resolution: %s (examples: 1080p, 720p, 480p, or 1920x1080)", opts.Resolution)
		}
	}

	return nil
}

func ResolveOutput(opts *Options) {
	if opts.Output != "" && opts.Format == "" {
		parts := strings.Split(opts.Output, ".")
		if len(parts) > 1 {
			opts.Format = parts[len(parts)-1]
		}
	}

	if opts.Format == "" {
		opts.Format = "mp4"
	}

	if opts.Output == "" {
		base := strings.TrimSuffix(opts.Input, "."+getExtension(opts.Input))
		opts.Output = base + "_converted." + opts.Format
	}

	if opts.Quality == "" {
		opts.Quality = QualityMedium
	}
}

func Convert(opts *Options, onProgress ProgressFunc) error {
	totalDuration, err := probeDuration(opts.Input)
	if err != nil {
		totalDuration = 0
	}

	args := buildFFmpegArgs(opts)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to capture ffmpeg output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	if onProgress != nil && totalDuration > 0 {
		parseProgress(stderr, totalDuration, onProgress)
	} else {
		io.Copy(io.Discard, stderr)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	return nil
}

func buildFFmpegArgs(opts *Options) []string {
	args := []string{"-i", opts.Input, "-y", "-progress", "pipe:2", "-nostats"}

	codec := "libx264"
	if opts.Codec != "" {
		codec = codecMap[opts.Codec]
	} else if opts.Format == "webm" {
		codec = "libvpx-vp9"
	}

	args = append(args, "-c:v", codec)

	crf := crfMap[opts.Quality]
	if strings.Contains(codec, "vpx") {
		args = append(args, "-crf", strconv.Itoa(crf), "-b:v", "0")
	} else {
		args = append(args, "-crf", strconv.Itoa(crf))
	}

	args = append(args, "-c:a", "aac", "-b:a", "128k")

	if opts.Resolution != "" {
		scale := resolveScale(opts.Resolution)
		args = append(args, "-vf", scale)
	}

	args = append(args, opts.Output)
	return args
}

func probeDuration(input string) (time.Duration, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

var timeRegex = regexp.MustCompile(`out_time_us=(\d+)`)

func parseProgress(r io.Reader, total time.Duration, onProgress ProgressFunc) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := timeRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			us, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				continue
			}
			current := time.Duration(us) * time.Microsecond
			percent := float64(current) / float64(total) * 100
			if percent > 100 {
				percent = 100
			}
			onProgress(percent)
		}
	}
}

func isValidResolution(res string) bool {
	presets := map[string]bool{
		"2160p": true, "1440p": true, "1080p": true,
		"720p": true, "480p": true, "360p": true,
	}
	if presets[res] {
		return true
	}
	matched, _ := regexp.MatchString(`^\d+x\d+$`, res)
	return matched
}

func resolveScale(res string) string {
	presets := map[string]string{
		"2160p": "scale=-2:2160",
		"1440p": "scale=-2:1440",
		"1080p": "scale=-2:1080",
		"720p":  "scale=-2:720",
		"480p":  "scale=-2:480",
		"360p":  "scale=-2:360",
	}
	if scale, ok := presets[res]; ok {
		return scale
	}
	parts := strings.Split(res, "x")
	return fmt.Sprintf("scale=%s:%s", parts[0], parts[1])
}

func getExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
