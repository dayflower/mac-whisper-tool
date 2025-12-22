package db

import (
	"time"
)

// RecordedMeeting represents a meeting recording metadata
type RecordedMeeting struct {
	ID                   []byte
	Date                 time.Time
	Title                *string
	BundleIdentifier     string
	AppName              string
	Duration             *float64
	DateCreated          time.Time
	DateUpdated          *time.Time
	MatchedCalendarTitle *string
}

// Session represents a transcription session
type Session struct {
	ID                []byte
	DateCreated       time.Time
	DateUpdated       *time.Time
	DateLastOpened    *time.Time
	TextPreview       *string
	AITitle           *string
	UserChosenTitle   *string
	DetectedLanguage  *string
	PlaybackDuration  *float64
	RecordedMeetingID []byte
	HasBeenDiarized   bool
}

// TranscriptLine represents a single line of transcribed text
type TranscriptLine struct {
	ID        []byte
	Text      string
	Start     int64 // milliseconds
	End       int64 // milliseconds
	SessionID []byte
	SpeakerID []byte
}

// Speaker represents a speaker in a diarized session
type Speaker struct {
	ID   []byte
	Name string
}

// MeetingInfo represents combined meeting and session information for listing
type MeetingInfo struct {
	SessionID         string
	DateCreated       time.Time // Original dateCreated from database
	DateStarted       time.Time // Estimated or same as DateCreated
	Duration          float64   // seconds
	Title             string
	Preview           string
	UseEstimatedStart bool
}

// TranscriptExport represents the export structure for transcripts
type TranscriptExport struct {
	Speaker   string
	Text      string
	SpeakedAt *time.Time // only for extended format
}

// FullExport represents the extended JSON export format
type FullExport struct {
	Title       string             `json:"title"`
	DateStarted *string            `json:"dateStarted,omitempty"`
	DateCreated string             `json:"dateCreated"`
	DateUpdated *string            `json:"dateUpdated,omitempty"`
	Duration    float64            `json:"duration"`
	Transcripts []TranscriptExport `json:"transcripts"`
}

// SimpleTranscriptExport represents the standard JSON export format (MacWhisper compatible)
type SimpleTranscriptExport struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
}
