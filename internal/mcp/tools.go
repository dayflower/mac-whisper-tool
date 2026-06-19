package mcp

import (
	"bytes"
	"context"
	"fmt"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/exporter"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchMeetingsInput defines the input schema for search_meetings tool
type SearchMeetingsInput struct {
	Keywords      []string `json:"keywords,omitempty" jsonschema:"Keywords to search in transcript content (AND condition)"`
	TitleKeywords []string `json:"titleKeywords,omitempty" jsonschema:"Keywords to search in meeting titles (AND condition)"`
	After         string   `json:"after,omitempty" jsonschema:"Filter meetings started after this datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)"`
	Before        string   `json:"before,omitempty" jsonschema:"Filter meetings started before this datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)"`
	Limit         int      `json:"limit,omitempty" jsonschema:"Maximum number of results (default: 10; max: 100)"`
	SessionType   string   `json:"sessionType,omitempty" jsonschema:"Filter by session source type: recorded-meeting, system-audio, voice-memo, podcast, youtube, download, imported (empty = all)"`
}

// HandleSearchMeetings handles the search_meetings tool request
func HandleSearchMeetings(database *db.DB, estimateStart bool, legacyMode bool, logFunc db.LogFunc) func(context.Context, *mcpsdk.CallToolRequest, SearchMeetingsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input SearchMeetingsInput) (*mcpsdk.CallToolResult, any, error) {
		logFunc("Search request: keywords=%v, titleKeywords=%v, after=%s, before=%s, limit=%d, sessionType=%s",
			input.Keywords, input.TitleKeywords, input.After, input.Before, input.Limit, input.SessionType)
		// Convert input to search params
		params := SearchMeetingsParams{
			Keywords:      input.Keywords,
			TitleKeywords: input.TitleKeywords,
			After:         input.After,
			Before:        input.Before,
			Limit:         input.Limit,
			SessionType:   input.SessionType,
		}

		// Perform search
		result, err := SearchMeetings(database, params, estimateStart, logFunc)
		if err != nil {
			// Return error as tool error content
			return &mcpsdk.CallToolResult{
				IsError: true,
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: fmt.Sprintf("Search failed: %v", err),
					},
				},
			}, nil, nil
		}

		// Build response content
		var content []mcpsdk.Content

		// 1. Add JSON-formatted summary
		jsonSummary, err := FormatSearchResultJSON(result)
		if err != nil {
			return &mcpsdk.CallToolResult{
				IsError: true,
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: fmt.Sprintf("Failed to format results: %v", err),
					},
				},
			}, nil, nil
		}
		content = append(content, &mcpsdk.TextContent{
			Text: jsonSummary,
		})

		// 2. Add resource links for each meeting (only in non-legacy mode)
		if !legacyMode {
			logFunc("Building %d resource links", len(result.Meetings))
			for _, meeting := range result.Meetings {
				uri := FormatSessionURI(meeting.SessionID)
				name := FormatResourceName(meeting.DateStarted, meeting.Title)

				content = append(content, &mcpsdk.ResourceLink{
					URI:      uri,
					Name:     name,
					MIMEType: "text/markdown",
				})
			}
		}

		return &mcpsdk.CallToolResult{
			Content: content,
			IsError: false,
		}, nil, nil
	}
}

// GetMeetingInput defines the input schema for get_meeting tool
type GetMeetingInput struct {
	SessionID string `json:"sessionId" jsonschema:"Session ID (base64 encoded)"`
}

// HandleGetMeeting handles the get_meeting tool request (legacy mode only)
func HandleGetMeeting(database *db.DB, estimateStart bool, logFunc db.LogFunc) func(context.Context, *mcpsdk.CallToolRequest, GetMeetingInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input GetMeetingInput) (*mcpsdk.CallToolResult, any, error) {
		logFunc("Get meeting request: sessionId=%s", input.SessionID)

		// Validate session ID
		if input.SessionID == "" {
			return &mcpsdk.CallToolResult{
				IsError: true,
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: "Session ID is required",
					},
				},
			}, nil, nil
		}

		// Get transcript lines
		logFunc("Fetching transcript lines for session: %s", input.SessionID)
		transcripts, info, err := database.GetTranscriptLines(input.SessionID, estimateStart)
		if err != nil {
			logFunc("Database error: %v", err)
			return &mcpsdk.CallToolResult{
				IsError: true,
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: fmt.Sprintf("Session not found or database error: %v", err),
					},
				},
			}, nil, nil
		}

		logFunc("Found %d transcript lines", len(transcripts))

		// Generate extended Markdown content
		var buf bytes.Buffer
		if err := exporter.ExportMarkdown(&buf, transcripts, true, true, info); err != nil {
			logFunc("Export failed: %v", err)
			return &mcpsdk.CallToolResult{
				IsError: true,
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: fmt.Sprintf("Failed to export markdown: %v", err),
					},
				},
			}, nil, nil
		}

		logFunc("Exported markdown: %d bytes", buf.Len())

		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{
					Text: buf.String(),
				},
			},
			IsError: false,
		}, nil, nil
	}
}

// RegisterTools registers all MCP tools
func RegisterTools(server *mcpsdk.Server, database *db.DB, estimateStart bool, legacyMode bool, logFunc db.LogFunc) {
	// Register search_meetings tool
	searchTool := &mcpsdk.Tool{
		Name:        "search_meetings",
		Description: "Search MacWhisper meeting transcriptions by content, title, or date range. Returns resource URIs for matching sessions.",
	}
	mcpsdk.AddTool(server, searchTool, HandleSearchMeetings(database, estimateStart, legacyMode, logFunc))

	// Register get_meeting tool (legacy mode only)
	if legacyMode {
		getMeetingTool := &mcpsdk.Tool{
			Name:        "get_meeting",
			Description: "Get meeting transcript by session ID. Returns full transcript in Markdown format.",
		}
		mcpsdk.AddTool(server, getMeetingTool, HandleGetMeeting(database, estimateStart, logFunc))
	}
}
