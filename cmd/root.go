package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	dbPath  string
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "mac-whisper-export",
	Short: "Export meeting transcriptions from MacWhisper's database",
	Long: `A CLI tool to export meeting transcriptions from MacWhisper's database.

This tool provides command-line access to MacWhisper's SQLite database,
enabling users to list and export meeting transcriptions in various formats
(Markdown, JSON).`,
	Version: "0.1.0",
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
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "~/Library/Application Support/MacWhisper/Database/main.sqlite", "Path to MacWhisper database file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output to stderr")
}

// logVerbose prints message to stderr if verbose mode is enabled
func logVerbose(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
