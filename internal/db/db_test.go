package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
)

func TestSessionTypeCondition(t *testing.T) {
	cases := map[string]bool{
		"":                 false, // no restriction, no error
		"recorded-meeting": false,
		"system-audio":     false,
		"voice-memo":       false,
		"podcast":          false,
		"youtube":          false,
		"download":         false,
		"imported":         false,
		"bogus":            true, // invalid -> error
	}
	for input, wantErr := range cases {
		cond, err := sessionTypeCondition(input)
		if wantErr {
			if err == nil {
				t.Errorf("sessionTypeCondition(%q): expected error, got nil", input)
			}
			continue
		}
		if err != nil {
			t.Errorf("sessionTypeCondition(%q): unexpected error: %v", input, err)
		}
		if input == "" && cond != "" {
			t.Errorf("sessionTypeCondition(\"\"): expected empty condition, got %q", cond)
		}
		if input != "" && cond == "" {
			t.Errorf("sessionTypeCondition(%q): expected non-empty condition", input)
		}
	}
}

// setupTestDB creates a temporary MacWhisper-like SQLite database with a minimal
// schema and one session per source type (plus a soft-deleted one), then returns
// the path to it.
func setupTestDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.sqlite")

	conn, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", path))
	if err != nil {
		t.Fatalf("open writable db: %v", err)
	}
	defer conn.Close()

	schema := `
		CREATE TABLE session (
			id BLOB PRIMARY KEY NOT NULL,
			dateCreated DATETIME NOT NULL,
			dateDeleted DOUBLE,
			playbackDuration DOUBLE,
			textPreview TEXT,
			fullText TEXT,
			aiTitle TEXT,
			userChosenTitle TEXT,
			originalFilename TEXT,
			isFromYoutube BOOLEAN NOT NULL DEFAULT 0,
			recordedMeetingID BLOB,
			systemAudioRecordingID BLOB,
			voiceMemoID BLOB,
			podcastID BLOB,
			downloadMetadataID BLOB
		);
		CREATE TABLE recordedmeeting (id BLOB PRIMARY KEY, title TEXT, matchedCalendarTitle TEXT, duration DOUBLE);
		CREATE TABLE systemaudiorecording (id BLOB PRIMARY KEY, title TEXT, duration DOUBLE);
		CREATE TABLE voicememos (id BLOB PRIMARY KEY, title TEXT);
		CREATE TABLE podcast (id BLOB PRIMARY KEY, title TEXT);
		CREATE TABLE downloadmetadata (id BLOB PRIMARY KEY, youtubeTitle TEXT);

		INSERT INTO recordedmeeting (id, title, duration) VALUES (x'01', 'Standup', 1800);
		INSERT INTO systemaudiorecording (id, title, duration) VALUES (x'02', 'System Audio Cap', 600);
		INSERT INTO voicememos (id, title) VALUES (x'03', 'Memo One');
		INSERT INTO podcast (id, title) VALUES (x'04', 'Episode 12');
		INSERT INTO downloadmetadata (id, youtubeTitle) VALUES (x'05', 'YT Clip');

		-- one per type, dateCreated ascending so ORDER BY DESC is deterministic
		INSERT INTO session (id, dateCreated, recordedMeetingID)      VALUES (x'a1', '2026-01-01 10:00:00', x'01');
		INSERT INTO session (id, dateCreated, systemAudioRecordingID) VALUES (x'a2', '2026-01-02 10:00:00', x'02');
		INSERT INTO session (id, dateCreated, voiceMemoID)            VALUES (x'a3', '2026-01-03 10:00:00', x'03');
		INSERT INTO session (id, dateCreated, podcastID)              VALUES (x'a4', '2026-01-04 10:00:00', x'04');
		INSERT INTO session (id, dateCreated, downloadMetadataID, isFromYoutube) VALUES (x'a5', '2026-01-05 10:00:00', x'05', 1);
		INSERT INTO session (id, dateCreated, originalFilename)       VALUES (x'a6', '2026-01-06 10:00:00', 'meeting.mp4');
		-- soft-deleted session must be excluded
		INSERT INTO session (id, dateCreated, originalFilename, dateDeleted) VALUES (x'a7', '2026-01-07 10:00:00', 'trashed.mp4', 123456.0);
	`
	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return path
}

func TestListMeetingsReturnsAllTypes(t *testing.T) {
	path := setupTestDB(t)
	database, err := Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	meetings, err := database.ListMeetings(ListMeetingsFilters{})
	if err != nil {
		t.Fatalf("ListMeetings: %v", err)
	}

	// 6 active sessions (the soft-deleted one excluded)
	if len(meetings) != 6 {
		t.Fatalf("expected 6 meetings, got %d", len(meetings))
	}

	gotTypes := make([]string, 0, len(meetings))
	for _, m := range meetings {
		gotTypes = append(gotTypes, m.Type)
	}
	sort.Strings(gotTypes)
	want := []string{"imported", "podcast", "recorded-meeting", "system-audio", "voice-memo", "youtube"}
	for i := range want {
		if gotTypes[i] != want[i] {
			t.Errorf("types mismatch: got %v, want %v", gotTypes, want)
			break
		}
	}

	// Titles resolved per type (recorded meeting from linked table, imported from filename)
	byType := map[string]MeetingInfo{}
	for _, m := range meetings {
		byType[m.Type] = m
	}
	if got := byType["recorded-meeting"].Title; got != "Standup" {
		t.Errorf("recorded-meeting title: got %q, want Standup", got)
	}
	if got := byType["imported"].Title; got != "meeting.mp4" {
		t.Errorf("imported title: got %q, want meeting.mp4", got)
	}
	// the download-metadata row has isFromYoutube=1, so it classifies as youtube
	if got := byType["youtube"].Title; got != "YT Clip" {
		t.Errorf("youtube title: got %q, want YT Clip", got)
	}
	// recorded meeting carries duration from linked table
	if got := byType["recorded-meeting"].Duration; got != 1800 {
		t.Errorf("recorded-meeting duration: got %v, want 1800", got)
	}
}

func TestListMeetingsSessionTypeFilter(t *testing.T) {
	path := setupTestDB(t)
	database, err := Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	for _, tc := range []struct {
		sessionType string
		wantCount   int
	}{
		{"imported", 1},
		{"recorded-meeting", 1},
		{"voice-memo", 1},
		{"youtube", 1},
		{"download", 0}, // the only download row is a youtube one
	} {
		meetings, err := database.ListMeetings(ListMeetingsFilters{SessionType: tc.sessionType})
		if err != nil {
			t.Fatalf("ListMeetings(type=%s): %v", tc.sessionType, err)
		}
		if len(meetings) != tc.wantCount {
			t.Errorf("type=%s: got %d, want %d", tc.sessionType, len(meetings), tc.wantCount)
		}
	}

	// invalid type surfaces an error
	if _, err := database.ListMeetings(ListMeetingsFilters{SessionType: "nope"}); err == nil {
		t.Error("expected error for invalid session type, got nil")
	}
}
