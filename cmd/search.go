package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/exporter"
	"github.com/dayflower/mac-whisper-tool/internal/formatter"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
	"github.com/spf13/cobra"
)

var (
	// Search command flags
	searchKeywords      []string
	searchTitleKeywords []string
	searchStartTime     string
	searchEndTime       string
	searchLimit         int
	searchFormat        string
	searchEstimateStart bool
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search meetings from MacWhisper database",
	Long: `Search meetings from MacWhisper database by keywords and filters.

Search supports filtering by:
- Content keywords (full text search)
- Title keywords
- Date range (after/before)

At least one search criterion must be specified.

Examples:
  # Search by content keyword
  mac-whisper-tool search -k "budget"

  # Search by multiple keywords (AND condition)
  mac-whisper-tool search -k "zoom" -k "meeting"

  # Search by title
  mac-whisper-tool search -t "standup"

  # Search with date range
  mac-whisper-tool search -k "Q4" -s 2025-01-01 -e 2025-03-31

  # Search with JSON output
  mac-whisper-tool search -k "project" -f json`,
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringArrayVarP(&searchKeywords, "keywords", "k", []string{},
		"Content keywords (repeatable, AND condition)")
	searchCmd.Flags().StringArrayVarP(&searchTitleKeywords, "title", "t", []string{},
		"Title keywords (repeatable, AND condition)")
	searchCmd.Flags().StringVarP(&searchStartTime, "start", "s", "",
		"Filter by start datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	searchCmd.Flags().StringVarP(&searchEndTime, "end", "e", "",
		"Filter by end datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 20,
		"Maximum number of meetings to display (negative for all)")
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "table",
		"Output format: table or json")
	searchCmd.Flags().BoolVar(&searchEstimateStart, "estimate-start", false,
		"Estimate meeting start time from dateCreated and duration")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// 1. Validate: at least one search criterion
	if len(searchKeywords) == 0 &&
		len(searchTitleKeywords) == 0 &&
		searchStartTime == "" &&
		searchEndTime == "" {
		return fmt.Errorf("at least one search criterion must be specified (--keywords, --title, --start, or --end)")
	}

	// 2. Open database
	logVerbose("Opening database: %s", dbPath)
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// 3. Parse time filters
	var startTime, endTime *time.Time
	if searchStartTime != "" {
		t, err := utils.ParseDateTime(searchStartTime)
		if err != nil {
			return err
		}
		startTime = &t
		logVerbose("Start time filter: %s", utils.FormatDateTime(t))
	}
	if searchEndTime != "" {
		t, err := utils.ParseDateTime(searchEndTime)
		if err != nil {
			return err
		}
		endTime = &t
		logVerbose("End time filter: %s", utils.FormatDateTime(t))
	}

	// 4. Set limit
	limit := searchLimit
	if limit < 0 {
		limit = 0
		logVerbose("Retrieving all matching meetings")
	} else {
		logVerbose("Retrieving up to %d meetings", limit)
	}

	// 5. Log search criteria
	if len(searchKeywords) > 0 {
		logVerbose("Content keywords: %v", searchKeywords)
	}
	if len(searchTitleKeywords) > 0 {
		logVerbose("Title keywords: %v", searchTitleKeywords)
	}

	// 6. Determine start time estimation
	estimateStart := searchEstimateStart
	if estimateStart {
		logVerbose("Estimating meeting start times")
	}

	// 7. Build filters
	filters := db.ListMeetingsFilters{
		StartTime:     startTime,
		EndTime:       endTime,
		Limit:         limit,
		EstimateStart: estimateStart,
		Keywords:      searchKeywords,
		TitleKeywords: searchTitleKeywords,
	}

	// 8. Query meetings
	logVerbose("Searching meetings...")
	meetings, err := database.ListMeetings(filters)
	if err != nil {
		return fmt.Errorf("failed to search meetings: %w", err)
	}

	logVerbose("Found %d matching meetings", len(meetings))

	// 9. Format output
	switch searchFormat {
	case "table":
		formatter.FormatMeetingTable(meetings)
	case "json":
		if err := exporter.ExportListJSON(os.Stdout, meetings); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid format: %s (must be 'table' or 'json')", searchFormat)
	}

	return nil
}
