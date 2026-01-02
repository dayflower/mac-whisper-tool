package mcp

import (
	"bytes"
	"context"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/exporter"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// HandleReadResource handles resources/read requests
func HandleReadResource(database *db.DB, estimateStart bool, logFunc db.LogFunc) mcpsdk.ResourceHandler {
	return func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		params := req.GetParams().(*mcpsdk.ReadResourceParams)
		uri := params.URI

		logFunc("Resource read request: %s", uri)

		// Parse URI to extract session ID
		sessionID, err := ParseSessionURI(uri)
		if err != nil {
			logFunc("URI parsing failed: %v", err)
			logFunc("Resource not found: %s", uri)
			return nil, mcpsdk.ResourceNotFoundError(uri)
		}

		logFunc("Parsed session ID: %s", sessionID)

		// Get transcript lines
		logFunc("Fetching transcript lines for session: %s", sessionID)
		transcripts, info, err := database.GetTranscriptLines(sessionID, estimateStart)
		if err != nil {
			// Session not found or database error
			logFunc("Database error: %v", err)
			logFunc("Resource not found: %s", uri)
			return nil, mcpsdk.ResourceNotFoundError(uri)
		}

		logFunc("Found %d transcript lines", len(transcripts))

		// Generate extended Markdown content
		var buf bytes.Buffer
		if err := exporter.ExportMarkdown(&buf, transcripts, true, true, info); err != nil {
			logFunc("Export failed: %v", err)
			logFunc("Resource not found: %s", uri)
			return nil, mcpsdk.ResourceNotFoundError(uri)
		}

		logFunc("Exported markdown: %d bytes", buf.Len())
		logFunc("Resource read successful")

		// Return resource contents
		return &mcpsdk.ReadResourceResult{
			Contents: []*mcpsdk.ResourceContents{{
				URI:      uri,
				MIMEType: "text/markdown",
				Text:     buf.String(),
			}},
		}, nil
	}
}

// RegisterResources registers resource templates
func RegisterResources(server *mcpsdk.Server, database *db.DB, estimateStart bool, logFunc db.LogFunc) {
	// Register resource template (not individual resources)
	template := &mcpsdk.ResourceTemplate{
		URITemplate: "macwhisper://localhost/session/{sessionID}",
		Name:        "MacWhisper Meeting Transcription",
		Description: "Access meeting transcription by session ID. Use search_meetings tool to find session IDs.",
		MIMEType:    "text/markdown",
	}

	server.AddResourceTemplate(template, HandleReadResource(database, estimateStart, logFunc))
}
