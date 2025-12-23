package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
)

// SearchMeetingsParams contains search parameters from MCP tool
type SearchMeetingsParams struct {
	Keywords      []string
	TitleKeywords []string
	After         string // YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS
	Before        string // YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS
	Limit         int
}

// SearchMeetingsResult contains search results
type SearchMeetingsResult struct {
	Meetings []db.MeetingInfo
	Count    int
}

// SearchMeetings searches meetings based on MCP tool parameters
func SearchMeetings(database *db.DB, params SearchMeetingsParams, estimateStart bool, logFunc db.LogFunc) (*SearchMeetingsResult, error) {
	// Parse date filters
	var startTime, endTime *time.Time

	if params.After != "" {
		t, err := utils.ParseDateTime(params.After)
		if err != nil {
			return nil, fmt.Errorf("invalid 'after' parameter: %w", err)
		}
		startTime = &t
		logFunc("Parsed 'after' filter: %v", startTime)
	}

	if params.Before != "" {
		t, err := utils.ParseDateTime(params.Before)
		if err != nil {
			return nil, fmt.Errorf("invalid 'before' parameter: %w", err)
		}
		endTime = &t
		logFunc("Parsed 'before' filter: %v", endTime)
	}

	// Validate and set limit
	limit := params.Limit
	if limit < 1 {
		limit = 10 // Default
	} else if limit > 100 {
		limit = 100 // Max
	}
	logFunc("Using limit: %d", limit)

	// Build filters
	filters := db.ListMeetingsFilters{
		StartTime:     startTime,
		EndTime:       endTime,
		Limit:         limit,
		EstimateStart: estimateStart,
		Keywords:      params.Keywords,
		TitleKeywords: params.TitleKeywords,
	}

	// Query database
	meetings, err := database.ListMeetings(filters, logFunc)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	logFunc("Search completed: found %d meetings", len(meetings))

	return &SearchMeetingsResult{
		Meetings: meetings,
		Count:    len(meetings),
	}, nil
}

// FormatSearchSummary formats search results as text summary
// Format per specification in MCP_PLAN.md
func FormatSearchSummary(result *SearchMeetingsResult) string {
	if result.Count == 0 {
		return "No matching meetings found."
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "Found %d matching meetings:\n\n", result.Count)

	for i, meeting := range result.Meetings {
		// Format: "1. 2025-12-15 14:00 - Zoom Meeting\n   Duration: 45m 30s\n   Preview: ..."
		datetime := meeting.DateStarted.Format("2006-01-02 15:04")
		duration := utils.FormatDuration(meeting.Duration)

		fmt.Fprintf(&builder, "%d. %s - %s\n", i+1, datetime, meeting.Title)
		fmt.Fprintf(&builder, "   Duration: %s\n", duration)
		fmt.Fprintf(&builder, "   Preview: %s\n", meeting.Preview)

		if i < result.Count-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
