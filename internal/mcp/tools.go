package mcp

import (
	"context"
	"fmt"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchMeetingsInput defines the input schema for search_meetings tool
type SearchMeetingsInput struct {
	Keywords      []string `json:"keywords,omitempty" jsonschema:"Keywords to search in transcript content (AND condition)"`
	TitleKeywords []string `json:"titleKeywords,omitempty" jsonschema:"Keywords to search in meeting titles (AND condition)"`
	After         string   `json:"after,omitempty" jsonschema:"Filter meetings started after this datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)"`
	Before        string   `json:"before,omitempty" jsonschema:"Filter meetings started before this datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)"`
	Limit         int      `json:"limit,omitempty" jsonschema:"Maximum number of results (default: 10; max: 100)"`
}

// HandleSearchMeetings handles the search_meetings tool request
func HandleSearchMeetings(database *db.DB, estimateStart bool, logFunc db.LogFunc) func(context.Context, *mcpsdk.CallToolRequest, SearchMeetingsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input SearchMeetingsInput) (*mcpsdk.CallToolResult, any, error) {
		logFunc("Search request: keywords=%v, titleKeywords=%v, after=%s, before=%s, limit=%d",
			input.Keywords, input.TitleKeywords, input.After, input.Before, input.Limit)
		// Convert input to search params
		params := SearchMeetingsParams{
			Keywords:      input.Keywords,
			TitleKeywords: input.TitleKeywords,
			After:         input.After,
			Before:        input.Before,
			Limit:         input.Limit,
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

		// 1. Add text summary
		summary := FormatSearchSummary(result)
		content = append(content, &mcpsdk.TextContent{
			Text: summary,
		})

		// 2. Add resource links for each meeting
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

		return &mcpsdk.CallToolResult{
			Content: content,
			IsError: false,
		}, nil, nil
	}
}

// RegisterTools registers all MCP tools
func RegisterTools(server *mcpsdk.Server, database *db.DB, estimateStart bool, logFunc db.LogFunc) {
	tool := &mcpsdk.Tool{
		Name:        "search_meetings",
		Description: "Search MacWhisper meeting transcriptions by content, title, or date range. Returns resource URIs for matching sessions.",
	}

	mcpsdk.AddTool(server, tool, HandleSearchMeetings(database, estimateStart, logFunc))
}
