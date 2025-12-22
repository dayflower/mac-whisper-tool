package utils

import (
	"fmt"
	"strings"
	"time"
)

// ParseDateTime parses datetime string in various formats
// Accepts: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, YYYY-MM-DDTHH:MM:SS.sss
func ParseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	// Try different formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000",
	}

	var t time.Time
	var err error

	for _, format := range formats {
		t, err = time.ParseInLocation(format, s, time.Local)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format: %s (expected YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS.sss)", s)
}

// FormatDateTime formats time to ISO 8601 without timezone
// Format: 2006-01-02T15:04:05.000
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000")
}

// FormatDuration formats duration in seconds to human-readable string
// Example: 3665.5 -> "1h 1m 5s"
func FormatDuration(seconds float64) string {
	d := time.Duration(seconds * float64(time.Second))

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}
