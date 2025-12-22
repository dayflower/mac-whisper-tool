package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/exporter"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
	"github.com/spf13/cobra"
)

var (
	// Export command flags
	exportFormat        string
	exportOutput        string
	exportOutputDir     string
	exportExtend        bool
	exportEstimateStart bool

	// Batch export flags
	exportStartTime string
	exportEndTime   string
	exportLimit     int
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [session-id]",
	Short: "Export meeting transcription",
	Long: `Export meeting transcription in various formats.

By default, exports to stdout in Markdown format. You can specify:
- Output format: markdown (md) or json
- Content type: normal (default) or extended (--extend)
- Output file or directory
- Single session by ID or batch export with filters`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "markdown", "Output format: markdown (md) or json")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default: stdout)")
	exportCmd.Flags().StringVarP(&exportOutputDir, "output-dir", "c", "", "Output directory (filename will be auto-generated)")
	exportCmd.Flags().BoolVarP(&exportExtend, "extend", "x", false, "Output extended content (timestamps, metadata)")
	exportCmd.Flags().BoolVar(&exportEstimateStart, "estimate-start", false, "Estimate meeting start time from dateCreated and duration")

	// Batch export flags
	exportCmd.Flags().StringVarP(&exportStartTime, "start", "s", "", "Filter by start datetime (for batch export)")
	exportCmd.Flags().StringVarP(&exportEndTime, "end", "e", "", "Filter by end datetime (for batch export)")
	exportCmd.Flags().IntVarP(&exportLimit, "limit", "n", 0, "Maximum number of meetings to export (for batch export, negative for all)")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Determine mode: single session or batch
	isBatchMode := len(args) == 0

	if isBatchMode {
		// Batch mode requires output-dir
		if exportOutputDir == "" {
			return fmt.Errorf("batch export requires --output-dir to be specified")
		}
		return runBatchExport()
	}

	// Single session mode
	sessionID := args[0]
	return runSingleExport(sessionID)
}

func runSingleExport(sessionID string) error {
	logVerbose("Opening database: %s", dbPath)

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	logVerbose("Retrieving transcript for session: %s", sessionID)

	// Determine if start time estimation is needed
	// --estimate-start is used for time calculation, --extend determines if timestamps are included in output
	estimateStart := exportEstimateStart || exportExtend
	if estimateStart {
		logVerbose("Estimating meeting start time")
	}

	// Get transcript lines (with timestamps if --extend is used)
	includeTimestamps := exportExtend
	transcripts, info, err := database.GetTranscriptLines(sessionID, estimateStart)
	if err != nil {
		return fmt.Errorf("failed to get transcript: %w", err)
	}

	logVerbose("Found %d transcript lines", len(transcripts))

	// Determine output writer
	var writer io.Writer
	var outputFile *os.File

	if exportOutputDir != "" {
		// Auto-generate filename
		filename := generateFilename(info)
		outputPath := filepath.Join(exportOutputDir, filename)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(exportOutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		logVerbose("Writing to file: %s", outputPath)
		outputFile, err = os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
		writer = outputFile
	} else if exportOutput != "" {
		logVerbose("Writing to file: %s", exportOutput)
		outputFile, err = os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
		writer = outputFile
	} else {
		writer = os.Stdout
	}

	// Export based on format and content type
	format := normalizeFormat(exportFormat)
	switch format {
	case "markdown":
		if err := exporter.ExportMarkdown(writer, transcripts, includeTimestamps, exportExtend, info); err != nil {
			return err
		}
	case "json":
		if exportExtend {
			// Extended JSON includes full metadata
			if err := exporter.ExportJSONFull(writer, transcripts, info); err != nil {
				return err
			}
		} else {
			// Normal JSON is simple array format
			if err := exporter.ExportJSON(writer, transcripts); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("invalid format: %s (must be 'markdown', 'md', or 'json')", exportFormat)
	}

	logVerbose("Export completed successfully")
	return nil
}

func runBatchExport() error {
	logVerbose("Opening database: %s", dbPath)

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Parse time filters
	var startTime, endTime *time.Time
	if exportStartTime != "" {
		t, err := utils.ParseDateTime(exportStartTime)
		if err != nil {
			return err
		}
		startTime = &t
		logVerbose("Start time filter: %s", utils.FormatDateTime(t))
	}
	if exportEndTime != "" {
		t, err := utils.ParseDateTime(exportEndTime)
		if err != nil {
			return err
		}
		endTime = &t
		logVerbose("End time filter: %s", utils.FormatDateTime(t))
	}

	// Set limit
	limit := exportLimit
	if limit < 0 {
		limit = 0
		logVerbose("Exporting all meetings")
	} else if limit > 0 {
		logVerbose("Exporting up to %d meetings", limit)
	}

	// Determine if start time estimation is needed
	estimateStart := exportEstimateStart || exportExtend
	if estimateStart {
		logVerbose("Estimating meeting start times")
	}

	// Query meetings (with timestamps if --extend is used)
	includeTimestamps := exportExtend
	logVerbose("Querying meetings...")
	meetings, err := database.ListMeetings(startTime, endTime, limit, estimateStart)
	if err != nil {
		return fmt.Errorf("failed to list meetings: %w", err)
	}

	logVerbose("Found %d meetings to export", len(meetings))

	// Create output directory
	if err := os.MkdirAll(exportOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Export each meeting
	for i, meeting := range meetings {
		logVerbose("[%d/%d] Exporting session: %s", i+1, len(meetings), meeting.SessionID)

		// Get transcript lines
		transcripts, info, err := database.GetTranscriptLines(meeting.SessionID, estimateStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting session %s: %v\n", meeting.SessionID, err)
			continue
		}

		// Generate filename
		filename := generateFilename(info)
		outputPath := filepath.Join(exportOutputDir, filename)

		// Create output file
		outputFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file %s: %v\n", outputPath, err)
			continue
		}

		// Export based on format and content type
		format := normalizeFormat(exportFormat)
		var exportErr error
		switch format {
		case "markdown":
			exportErr = exporter.ExportMarkdown(outputFile, transcripts, includeTimestamps, exportExtend, info)
		case "json":
			if exportExtend {
				exportErr = exporter.ExportJSONFull(outputFile, transcripts, info)
			} else {
				exportErr = exporter.ExportJSON(outputFile, transcripts)
			}
		default:
			exportErr = fmt.Errorf("invalid format: %s", exportFormat)
		}

		outputFile.Close()

		if exportErr != nil {
			fmt.Fprintf(os.Stderr, "Error exporting session %s: %v\n", meeting.SessionID, exportErr)
			continue
		}

		logVerbose("  Exported to: %s", outputPath)
	}

	logVerbose("Batch export completed")
	return nil
}

// normalizeFormat normalizes format string (md -> markdown, etc.)
func normalizeFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "md" {
		return "markdown"
	}
	return format
}

// generateFilename generates a filename based on meeting info
// Format: "YYYY-MM-DDTHH:MM:SS.sss Title.ext"
func generateFilename(info *db.MeetingInfo) string {
	// Sanitize title for filename
	title := info.Title
	// Replace invalid filename characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		title = strings.ReplaceAll(title, char, "_")
	}

	// Determine file extension based on format
	ext := ".md"
	format := normalizeFormat(exportFormat)
	if format == "json" {
		ext = ".json"
	}

	// Format: "datetime title.ext"
	datetime := utils.FormatDateTime(info.DateStarted)
	return fmt.Sprintf("%s %s%s", datetime, title, ext)
}
