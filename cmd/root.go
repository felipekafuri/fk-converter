package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fk-converter",
	Short: "A fast video converter powered by ffmpeg",
	Long:  "fk-converter converts video files between formats with quality control.\nIt wraps ffmpeg with sensible defaults and a progress bar.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
