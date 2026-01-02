package mcp

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dayflower/mac-whisper-tool/internal/utils"
)

const (
	// MacWhisperScheme is the URI scheme for MacWhisper resources
	MacWhisperScheme = "macwhisper"

	// MacWhisperHost is the host for MacWhisper URIs (localhost for local resources)
	MacWhisperHost = "localhost"

	// SessionPathPrefix is the path prefix for session resources
	SessionPathPrefix = "/session/"
)

// ParseSessionURI parses a MacWhisper session URI and returns the session ID
// Format: macwhisper://localhost/session/{sessionID}
// Returns: sessionID (base64 encoded), error
func ParseSessionURI(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("invalid URI: %w", err)
	}

	if parsed.Scheme != MacWhisperScheme {
		return "", fmt.Errorf("invalid URI scheme: %s (expected %s)",
			parsed.Scheme, MacWhisperScheme)
	}

	if parsed.Host != MacWhisperHost {
		return "", fmt.Errorf("invalid URI host: %s (expected %s)",
			parsed.Host, MacWhisperHost)
	}

	if !strings.HasPrefix(parsed.Path, SessionPathPrefix) {
		return "", fmt.Errorf("invalid URI path: %s (expected /session/...)",
			parsed.Path)
	}

	sessionID := strings.TrimPrefix(parsed.Path, SessionPathPrefix)
	if sessionID == "" {
		return "", fmt.Errorf("session ID is empty")
	}

	// Validate session ID format (base64)
	if _, err := utils.DecodeSessionID(sessionID); err != nil {
		return "", fmt.Errorf("invalid session ID: %w", err)
	}

	return sessionID, nil
}

// FormatSessionURI generates a MacWhisper session URI from a session ID
// Input: sessionID (base64 encoded)
// Output: macwhisper://localhost/session/{sessionID}
func FormatSessionURI(sessionID string) string {
	return fmt.Sprintf("%s://%s%s%s", MacWhisperScheme, MacWhisperHost, SessionPathPrefix, sessionID)
}

// FormatResourceName generates a resource name for display
// Format: YYYY-MM-DD HH:MM - Title
func FormatResourceName(dateStarted time.Time, title string) string {
	// Format datetime without seconds and milliseconds
	datetime := dateStarted.Format("2006-01-02 15:04")
	return fmt.Sprintf("%s - %s", datetime, title)
}
