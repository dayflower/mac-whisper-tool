package utils

import (
	"unicode/utf8"
)

// TruncateString truncates a string to maxLen runes (not bytes),
// ensuring we don't break multibyte UTF-8 characters.
// Adds "..." suffix if truncated.
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	// Count runes, not bytes
	runeCount := utf8.RuneCountInString(s)
	if runeCount <= maxLen {
		return s
	}

	// Need to truncate
	// Reserve 3 characters for "..."
	targetLen := maxLen - 3
	if targetLen < 0 {
		targetLen = 0
	}

	// Convert to runes to handle multibyte characters properly
	runes := []rune(s)
	if len(runes) <= targetLen {
		return string(runes)
	}

	return string(runes[:targetLen]) + "..."
}
