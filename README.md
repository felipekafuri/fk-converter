# fk-converter

A fast CLI video converter powered by ffmpeg. Convert formats, control quality, and resize videos from your terminal.

## Install

```bash
go install github.com/felipekafuri/fk-converter@latest
```

Requires [ffmpeg](https://ffmpeg.org/) installed on your system.

## Usage

```bash
# Basic format conversion
fk-converter convert video.mov -o output.mp4

# Compress with low quality (smaller file)
fk-converter convert video.mp4 -q low -o compressed.mp4

# Convert to WebM with resolution downscale
fk-converter convert video.mp4 -f webm -r 720p

# High quality H.265 encoding
fk-converter convert video.mov --codec h265 -q high -o output.mp4
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output file path (auto-generated if omitted) |
| `--format` | `-f` | Output format: `mp4`, `mkv`, `webm`, `avi`, `mov` |
| `--quality` | `-q` | Quality preset: `low`, `medium`, `high`, `lossless` (default: `medium`) |
| `--resolution` | `-r` | Target resolution: `2160p`, `1080p`, `720p`, `480p`, `360p` |
| `--codec` | | Video codec: `h264`, `h265`, `vp9` |

## Quality Presets

| Preset | CRF | Use case |
|--------|-----|----------|
| `low` | 28 | Small files, sharing |
| `medium` | 23 | Balanced (default) |
| `high` | 18 | High quality, larger files |
| `lossless` | 0 | No quality loss |

## License

MIT
