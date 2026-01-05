package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dayflower/mac-whisper-tool/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpEstimateStart bool
	mcpLegacyMode    bool
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI assistant integration",
	Long: `Start an MCP (Model Context Protocol) server that enables AI assistants
to search and retrieve meeting transcriptions from MacWhisper's database.

The server communicates via stdio using JSON-RPC 2.0 protocol.

Example usage in Claude Desktop configuration:
  {
    "mcpServers": {
      "mac-whisper": {
        "command": "/usr/local/bin/mac-whisper-tool",
        "args": ["mcp"]
      }
    }
  }
`,
	RunE: runMCP,
}

func init() {
	// MCP command uses global flags (dbPath, verbose) from root.go
	mcpCmd.Flags().BoolVar(&mcpEstimateStart, "estimate-start", false, "Estimate meeting start time by subtracting duration from creation time")
	mcpCmd.Flags().BoolVar(&mcpLegacyMode, "legacy", false, "Enable legacy mode for MCP clients that don't support Resources")
}

func runMCP(cmd *cobra.Command, args []string) error {
	logVerbose("Starting MCP server")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logVerbose("Received signal: %v", sig)
		logVerbose("Shutting down MCP server...")
		cancel()
	}()

	// Run MCP server
	config := mcp.ServerConfig{
		DBPath:        dbPath,
		Verbose:       verbose,
		EstimateStart: mcpEstimateStart,
		LegacyMode:    mcpLegacyMode,
	}

	if err := mcp.RunServer(ctx, config); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}
