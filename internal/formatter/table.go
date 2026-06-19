package formatter

import (
	"os"

	"github.com/dayflower/mac-whisper-tool/internal/db"
	"github.com/dayflower/mac-whisper-tool/internal/utils"
	"github.com/olekukonko/tablewriter"
)

// FormatMeetingTable formats meetings as a table and writes to stdout
func FormatMeetingTable(meetings []db.MeetingInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Session ID", "Start Time", "Duration", "Type", "Title", "Preview"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, m := range meetings {
		table.Append([]string{
			m.SessionID,
			utils.FormatDateTime(m.DateStarted),
			utils.FormatDuration(m.Duration),
			m.Type,
			m.Title,
			m.Preview,
		})
	}

	table.Render()
}
