package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/felipekafuri/fk-converter/converter"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	output     string
	format     string
	quality    string
	resolution string
	codec      string
)

var convertCmd = &cobra.Command{
	Use:   "convert <input-file>",
	Short: "Convert a video file",
	Long: `Convert a video file to a different format and/or quality.

Examples:
  fk-converter convert video.mov -o output.mp4
  fk-converter convert video.avi -f mkv -q high
  fk-converter convert video.mp4 -r 720p -q low
  fk-converter convert video.mov --codec h265 -q high -o compressed.mp4`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := converter.CheckFFmpeg(); err != nil {
			return err
		}

		opts := &converter.Options{
			Input:      args[0],
			Output:     output,
			Format:     format,
			Quality:    converter.Quality(quality),
			Resolution: resolution,
			Codec:      codec,
		}

		converter.ResolveOutput(opts)

		if err := converter.ValidateOptions(opts); err != nil {
			return err
		}

		fmt.Printf("Converting: %s → %s\n", opts.Input, opts.Output)
		fmt.Printf("Format: %s | Quality: %s", opts.Format, opts.Quality)
		if opts.Resolution != "" {
			fmt.Printf(" | Resolution: %s", opts.Resolution)
		}
		if opts.Codec != "" {
			fmt.Printf(" | Codec: %s", opts.Codec)
		}
		fmt.Println()

		bar := progressbar.NewOptions(100,
			progressbar.OptionSetDescription("Converting"),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowBytes(false),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionClearOnFinish(),
		)

		start := time.Now()

		err := converter.Convert(opts, func(percent float64) {
			bar.Set(int(percent))
		})
		if err != nil {
			fmt.Fprintln(os.Stderr)
			return err
		}

		bar.Finish()
		elapsed := time.Since(start).Round(time.Millisecond)

		info, _ := os.Stat(opts.Output)
		size := ""
		if info != nil {
			mb := float64(info.Size()) / 1024 / 1024
			size = fmt.Sprintf(" (%.1f MB)", mb)
		}

		fmt.Printf("\nDone in %s → %s%s\n", elapsed, opts.Output, size)
		return nil
	},
}

func init() {
	convertCmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	convertCmd.Flags().StringVarP(&format, "format", "f", "", "Output format (mp4, mkv, webm, avi, mov)")
	convertCmd.Flags().StringVarP(&quality, "quality", "q", "", "Quality preset: low, medium, high, lossless (default: medium)")
	convertCmd.Flags().StringVarP(&resolution, "resolution", "r", "", "Target resolution (e.g. 1080p, 720p, 480p)")
	convertCmd.Flags().StringVar(&codec, "codec", "", "Video codec (h264, h265, vp9)")

	rootCmd.AddCommand(convertCmd)
}
