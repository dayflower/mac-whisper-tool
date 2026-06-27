package cmd

import (
	"fmt"
	"os"

	"github.com/dayflower/mac-whisper-tool/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	dbPath  string
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "mac-whisper-tool",
	Short: "CLI tool to manage MacWhisper meeting transcriptions",
	Long: `A CLI tool to search, list, and export meeting transcriptions from MacWhisper's database.

This tool provides command-line access to MacWhisper's SQLite database,
enabling users to:
- List meetings with optional filtering
- Search meetings by keywords and date ranges
- Export transcriptions in various formats (Markdown, JSON)
- Start an MCP server for AI assistant integration
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version string) {
	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Get default DB path with priority:
	// 1. --db flag (handled by cobra)
	// 2. MAC_WHISPER_DB environment variable
	// 3. Config files
	// 4. Default path
	defaultDBPath := config.GetDatabasePath()

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", defaultDBPath, "Path to MacWhisper database file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output to stderr")

	cobra.EnableCommandSorting = false

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(mcpCmd)
}

// logVerbose prints message to stderr if verbose mode is enabled
func logVerbose(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
