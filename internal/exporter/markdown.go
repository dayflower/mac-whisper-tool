package exporter

import (
	"fmt"
	"io"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
)

// ExportMarkdown exports transcripts in Markdown format
func ExportMarkdown(w io.Writer, transcripts []db.TranscriptExport, includeTimestamps bool, includeMetadata bool, info *db.MeetingInfo) error {
	// Write metadata header if extended mode
	if includeMetadata && info != nil {
		header := fmt.Sprintf("# %s\n\n", info.Title)
		if _, err := w.Write([]byte(header)); err != nil {
			return fmt.Errorf("failed to write markdown header: %w", err)
		}

		metadata := ""
		if info.UseEstimatedStart {
			metadata += fmt.Sprintf("- **Date Started**: %s\n", utils.FormatDateTime(info.DateStarted))
		}
		metadata += fmt.Sprintf("- **Date Created**: %s\n", utils.FormatDateTime(info.DateCreated))
		metadata += fmt.Sprintf("- **Duration**: %s\n\n", utils.FormatDuration(info.Duration))

		if _, err := w.Write([]byte(metadata)); err != nil {
			return fmt.Errorf("failed to write markdown metadata: %w", err)
		}
	}

	// Write transcripts
	for _, t := range transcripts {
		var line string
		if includeTimestamps && t.SpeakedAt != nil {
			line = fmt.Sprintf("- %s **%s**: %s\n", utils.FormatDateTime(*t.SpeakedAt), t.Speaker, t.Text)
		} else {
			line = fmt.Sprintf("- **%s**: %s\n", t.Speaker, t.Text)
		}

		if _, err := w.Write([]byte(line)); err != nil {
			return fmt.Errorf("failed to write markdown: %w", err)
		}
	}
	return nil
}
