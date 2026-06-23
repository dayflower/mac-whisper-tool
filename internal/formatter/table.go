package formatter

import (
	"os"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// FormatMeetingTable formats meetings as a table and writes to stdout.
func FormatMeetingTable(meetings []db.MeetingInfo) error {
	// Borderless, separator-less, left-aligned compact layout
	// (replicates the v0.0.5 NoWhiteSpace + tab padding style).
	cellCfg := tw.CellConfig{
		Formatting: tw.CellFormatting{AutoWrap: tw.WrapNone},
		Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
		Padding:    tw.CellPadding{Global: tw.Padding{Left: "", Right: "\t", Overwrite: true}},
	}
	headerCfg := cellCfg
	headerCfg.Formatting.AutoFormat = tw.On // title-case headers

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint()),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.SeparatorsNone,
				Lines:      tw.LinesNone,
			},
		}),
		tablewriter.WithConfig(tablewriter.Config{
			Header: headerCfg,
			Row:    cellCfg,
		}),
	)

	table.Header("Session ID", "Start Time", "Duration", "Title", "Preview")

	for _, m := range meetings {
		if err := table.Append(
			m.SessionID,
			utils.FormatDateTime(m.DateStarted),
			utils.FormatDuration(m.Duration),
			m.Title,
			m.Preview,
		); err != nil {
			return err
		}
	}

	return table.Render()
}
