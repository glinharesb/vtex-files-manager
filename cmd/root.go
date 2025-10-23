package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool

	// Build-time variables set via ldflags
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "vfm",
	Short: "VTEX Files Manager - Upload and manage files in VTEX",
	Long: `vfm (VTEX Files Manager) - A powerful CLI tool built in Go to upload
and manage images and files in VTEX accounts using VTEX CLI session.

Authentication:
  This tool reads your VTEX session from VTEX CLI.
  Make sure you're logged in by running 'vtex login' first.

Supported file types:
  - Images (Universal): jpg, jpeg, png, gif, svg, webp
  - Images (CMS only): bmp
  - Documents (CMS only): pdf, txt, json, xml
  - Web (CMS only): css, js

Maximum file size: 5MB per file`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
