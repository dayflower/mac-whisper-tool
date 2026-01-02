package mcp

import (
	"context"
	"fmt"
	"os"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerConfig contains MCP server configuration
type ServerConfig struct {
	DBPath        string
	Verbose       bool
	EstimateStart bool
}

// RunServer starts the MCP server with stdio transport
func RunServer(ctx context.Context, config ServerConfig) error {
	// Log to stderr (stdout is used by MCP protocol)
	logInfo := func(format string, args ...interface{}) {
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "[MCP] "+format+"\n", args...)
		}
	}

	logInfo("Starting MCP server (version: 0.2.0)")
	logInfo("Opening database: %s", config.DBPath)

	// Open database
	database, err := db.Open(config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	logInfo("Database opened successfully")

	// Create MCP server
	implementation := &mcpsdk.Implementation{
		Name:    "mac-whisper-tool",
		Version: "0.2.0",
	}

	server := mcpsdk.NewServer(implementation, nil)

	logInfo("Registering tools and resources...")

	// Register tools
	RegisterTools(server, database, config.EstimateStart, logInfo)

	// Register resources
	RegisterResources(server, database, config.EstimateStart, logInfo)

	logInfo("MCP server initialized")
	logInfo("Listening on stdio...")

	// Start server with stdio transport
	transport := &mcpsdk.StdioTransport{}
	if err := server.Run(ctx, transport); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	logInfo("MCP server stopped")
	return nil
}
