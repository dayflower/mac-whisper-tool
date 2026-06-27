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
	// List command flags
	listStartTime     string
	listEndTime       string
	listLimit         int
	listFormat        string
	listExtend        bool
	listEstimateStart bool
	listSessionType   string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List meetings from MacWhisper database",
	Long: `List meetings from MacWhisper database with optional filtering.

By default, displays the most recent 20 meetings in table format.
You can filter by date range, change the limit, or output as JSON.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listStartTime, "start", "s", "", "Filter by start datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	listCmd.Flags().StringVarP(&listEndTime, "end", "e", "", "Filter by end datetime (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "Maximum number of meetings to display (negative for all)")
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "Output format: table or json")
	listCmd.Flags().BoolVarP(&listExtend, "extend", "x", false, "Output extended content")
	listCmd.Flags().BoolVar(&listEstimateStart, "estimate-start", false, "Estimate meeting start time from dateCreated and duration")
	listCmd.Flags().StringVar(&listSessionType, "type", "", "Filter by session type: recorded-meeting, system-audio, voice-memo, podcast, youtube, download, imported (empty = all)")
}

func runList(cmd *cobra.Command, args []string) error {
	logVerbose("Opening database: %s", dbPath)

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Parse time filters
	var startTime, endTime *time.Time
	if listStartTime != "" {
		t, err := utils.ParseDateTime(listStartTime)
		if err != nil {
			return err
		}
		startTime = &t
		logVerbose("Start time filter: %s", utils.FormatDateTime(t))
	}
	if listEndTime != "" {
		t, err := utils.ParseDateTime(listEndTime)
		if err != nil {
			return err
		}
		endTime = &t
		logVerbose("End time filter: %s", utils.FormatDateTime(t))
	}

	// Set limit (-1 for unlimited)
	limit := listLimit
	if limit < 0 {
		limit = 0
		logVerbose("Retrieving all meetings")
	} else {
		logVerbose("Retrieving up to %d meetings", limit)
	}

	// Determine if start time estimation is needed
	estimateStart := listEstimateStart
	if estimateStart {
		logVerbose("Estimating meeting start times")
	}

	// Build filters
	filters := db.ListMeetingsFilters{
		StartTime:     startTime,
		EndTime:       endTime,
		Limit:         limit,
		EstimateStart: estimateStart,
		SessionType:   listSessionType,
	}

	// Query meetings
	logVerbose("Querying meetings...")
	meetings, err := database.ListMeetings(filters)
	if err != nil {
		return fmt.Errorf("failed to list meetings: %w", err)
	}

	logVerbose("Found %d meetings", len(meetings))

	// Format output
	switch listFormat {
	case "table":
		if err := formatter.FormatMeetingTable(meetings); err != nil {
			return err
		}
	case "json":
		if err := exporter.ExportListJSON(os.Stdout, meetings); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid format: %s (must be 'table' or 'json')", listFormat)
	}

	return nil
}
