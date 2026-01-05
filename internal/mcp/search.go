package mcp

import (
	"encoding/json"
	"fmt"
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

// SearchResultItem represents a single meeting in JSON format
type SearchResultItem struct {
	SessionID   string `json:"sessionId"`
	DateStarted string `json:"dateStarted"`
	Duration    string `json:"duration"`
	Title       string `json:"title"`
	Preview     string `json:"preview"`
}

// SearchResultJSON represents the JSON response structure
type SearchResultJSON struct {
	Message string             `json:"message"`
	Total   int                `json:"total"`
	Items   []SearchResultItem `json:"items"`
}

// FormatSearchResultJSON formats search results as JSON
func FormatSearchResultJSON(result *SearchMeetingsResult) (string, error) {
	message := fmt.Sprintf("Found %d matching meetings", result.Count)
	if result.Count == 0 {
		message = "No matching meetings found"
	}

	items := make([]SearchResultItem, 0, len(result.Meetings))
	for _, meeting := range result.Meetings {
		items = append(items, SearchResultItem{
			DateStarted: utils.FormatDateTime(meeting.DateStarted),
			Duration:    utils.FormatDuration(meeting.Duration),
			Title:       meeting.Title,
			SessionID:   meeting.SessionID,
			Preview:     meeting.Preview,
		})
	}

	response := SearchResultJSON{
		Message: message,
		Total:   result.Count,
		Items:   items,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}
