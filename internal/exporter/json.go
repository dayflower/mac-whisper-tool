package exporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
)

// ExportJSON exports transcripts in simple JSON format (MacWhisper compatible)
func ExportJSON(w io.Writer, transcripts []db.TranscriptExport) error {
	simpleTranscripts := make([]db.SimpleTranscriptExport, len(transcripts))
	for i, t := range transcripts {
		simpleTranscripts[i] = db.SimpleTranscriptExport{
			Speaker: t.Speaker,
			Text:    t.Text,
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(simpleTranscripts); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// ExportJSONFull exports transcripts in extended JSON format with full metadata
func ExportJSONFull(w io.Writer, transcripts []db.TranscriptExport, info *db.MeetingInfo) error {
	// Build full export structure
	fullExport := db.FullExport{
		Title:       info.Title,
		DateCreated: utils.FormatDateTime(info.DateCreated),
		Duration:    info.Duration,
		Transcripts: transcripts,
	}

	// Add optional fields if available
	if info.UseEstimatedStart {
		dateStarted := utils.FormatDateTime(info.DateStarted)
		fullExport.DateStarted = &dateStarted
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(fullExport); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// ExportListJSON exports meeting list in JSON format
func ExportListJSON(w io.Writer, meetings []db.MeetingInfo) error {
	type MeetingJSON struct {
		SessionID   string  `json:"sessionId"`
		DateStarted string  `json:"dateStarted"`
		Duration    float64 `json:"duration"`
		Type        string  `json:"type"`
		Title       string  `json:"title"`
		Preview     string  `json:"preview"`
	}

	jsonMeetings := make([]MeetingJSON, len(meetings))
	for i, m := range meetings {
		jsonMeetings[i] = MeetingJSON{
			SessionID:   m.SessionID,
			DateStarted: utils.FormatDateTime(m.DateStarted),
			Duration:    m.Duration,
			Type:        m.Type,
			Title:       m.Title,
			Preview:     m.Preview,
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonMeetings); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
