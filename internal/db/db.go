package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dayflower/mac-whisper-tool/internal/utils"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// LogFunc is a function type for logging messages (optional parameter)
type LogFunc func(format string, args ...interface{})

// DB represents a database connection
type DB struct {
	conn *sql.DB
}

// Open opens a connection to the MacWhisper database
func Open(dbPath string) (*DB, error) {
	// Expand ~ to home directory
	if dbPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[2:])
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", dbPath)
	}

	// Open database in readonly mode
	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Session source tables joined to resolve type, title and duration regardless of
// how the session was created (recorded meeting, system audio, voice memo, podcast,
// downloaded/YouTube media, or a plain imported file).
const sessionSourceJoins = `
	LEFT JOIN recordedmeeting rm ON s.recordedMeetingID = rm.id
	LEFT JOIN systemaudiorecording sa ON s.systemAudioRecordingID = sa.id
	LEFT JOIN voicememos vm ON s.voiceMemoID = vm.id
	LEFT JOIN podcast pc ON s.podcastID = pc.id
	LEFT JOIN downloadmetadata dm ON s.downloadMetadataID = dm.id`

// sessionTitleExpr resolves a human-readable title with a per-type fallback chain.
const sessionTitleExpr = `COALESCE(s.userChosenTitle, s.aiTitle, rm.title, rm.matchedCalendarTitle, sa.title, vm.title, pc.title, dm.youtubeTitle, s.originalFilename, 'Untitled')`

// sessionDurationExpr is the source-recording duration. It is NULL for types that
// have no own duration column; callers fall back to session.playbackDuration.
const sessionDurationExpr = `COALESCE(rm.duration, sa.duration)`

// sessionTypeExpr classifies a session by which source FK is set (at most one is).
const sessionTypeExpr = `CASE
	WHEN s.recordedMeetingID IS NOT NULL THEN 'recorded-meeting'
	WHEN s.systemAudioRecordingID IS NOT NULL THEN 'system-audio'
	WHEN s.voiceMemoID IS NOT NULL THEN 'voice-memo'
	WHEN s.podcastID IS NOT NULL THEN 'podcast'
	WHEN s.downloadMetadataID IS NOT NULL AND s.isFromYoutube = 1 THEN 'youtube'
	WHEN s.downloadMetadataID IS NOT NULL THEN 'download'
	ELSE 'imported'
END`

// ValidSessionTypes lists the session type labels accepted by the --type filter.
var ValidSessionTypes = []string{"recorded-meeting", "system-audio", "voice-memo", "podcast", "youtube", "download", "imported"}

// sessionTypeCondition returns a SQL predicate (without leading AND) restricting
// results to the given session type. An empty string means "no restriction".
func sessionTypeCondition(t string) (string, error) {
	switch t {
	case "":
		return "", nil
	case "recorded-meeting":
		return "s.recordedMeetingID IS NOT NULL", nil
	case "system-audio":
		return "s.systemAudioRecordingID IS NOT NULL", nil
	case "voice-memo":
		return "s.voiceMemoID IS NOT NULL", nil
	case "podcast":
		return "s.podcastID IS NOT NULL", nil
	case "youtube":
		return "s.downloadMetadataID IS NOT NULL AND s.isFromYoutube = 1", nil
	case "download":
		return "s.downloadMetadataID IS NOT NULL AND (s.isFromYoutube = 0 OR s.isFromYoutube IS NULL)", nil
	case "imported":
		return "s.recordedMeetingID IS NULL AND s.systemAudioRecordingID IS NULL AND s.voiceMemoID IS NULL AND s.podcastID IS NULL AND s.downloadMetadataID IS NULL", nil
	default:
		return "", fmt.Errorf("invalid session type %q (valid: %s)", t, strings.Join(ValidSessionTypes, ", "))
	}
}

// ListMeetings retrieves a list of meetings based on filters
// Optional logFunc parameter can be provided for debug logging
func (db *DB) ListMeetings(filters ListMeetingsFilters, logFunc ...LogFunc) ([]MeetingInfo, error) {
	// Helper for conditional logging
	log := func(format string, args ...interface{}) {
		if len(logFunc) > 0 && logFunc[0] != nil {
			logFunc[0](format, args...)
		}
	}

	var query string
	args := []interface{}{}

	// When EstimateStart is enabled, we need to filter by estimated start time
	// This requires a subquery to calculate the estimated time first
	if filters.EstimateStart && (filters.StartTime != nil || filters.EndTime != nil) {
		query = fmt.Sprintf(`
			SELECT
				s.id,
				s.dateCreated,
				s.playbackDuration,
				%[1]s AS title,
				COALESCE(s.textPreview, ''),
				%[2]s AS recordingDuration,
				%[3]s AS sessionType,
				CASE
					WHEN COALESCE(%[2]s, s.playbackDuration, 0) > 0 THEN
						datetime(s.dateCreated, '-' || CAST(COALESCE(%[2]s, s.playbackDuration) AS INTEGER) || ' seconds')
					ELSE
						s.dateCreated
				END AS estimatedStart
			FROM session s
			%[4]s
			WHERE s.dateDeleted IS NULL
		`, sessionTitleExpr, sessionDurationExpr, sessionTypeExpr, sessionSourceJoins)
	} else {
		query = fmt.Sprintf(`
			SELECT
				s.id,
				s.dateCreated,
				s.playbackDuration,
				%[1]s AS title,
				COALESCE(s.textPreview, ''),
				%[2]s AS recordingDuration,
				%[3]s AS sessionType
			FROM session s
			%[4]s
			WHERE s.dateDeleted IS NULL
		`, sessionTitleExpr, sessionDurationExpr, sessionTypeExpr, sessionSourceJoins)
	}

	// Restrict to a specific session type if requested (e.g. --type imported)
	typeCond, err := sessionTypeCondition(filters.SessionType)
	if err != nil {
		return nil, err
	}
	if typeCond != "" {
		query += " AND " + typeCond
	}

	// Add content keyword filters (AND condition)
	for _, keyword := range filters.Keywords {
		query += " AND s.fullText LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// Add title keyword filters (AND condition)
	for _, keyword := range filters.TitleKeywords {
		query += " AND title LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// Add time filters
	// When EstimateStart is enabled, filter by estimated start time; otherwise by dateCreated
	if filters.StartTime != nil {
		if filters.EstimateStart {
			query += " AND estimatedStart >= ?"
		} else {
			query += " AND s.dateCreated >= ?"
		}
		args = append(args, filters.StartTime.Format("2006-01-02 15:04:05"))
	}
	if filters.EndTime != nil {
		if filters.EstimateStart {
			query += " AND estimatedStart <= ?"
		} else {
			query += " AND s.dateCreated <= ?"
		}
		args = append(args, filters.EndTime.Format("2006-01-02 15:04:05"))
	}

	query += " ORDER BY s.dateCreated DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	// Log query information
	log("SQL Query: %s", query)
	log("SQL Args: %v", args)

	// Get total session count for debugging
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM session s WHERE s.dateDeleted IS NULL"
	if err := db.conn.QueryRow(countQuery).Scan(&totalCount); err == nil {
		log("Database total sessions: %d", totalCount)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query meetings: %w", err)
	}
	defer rows.Close()

	var meetings []MeetingInfo
	for rows.Next() {
		var (
			id                []byte
			dateCreated       time.Time
			playbackDuration  *float64
			title             string
			preview           string
			recordingDuration *float64
			sessionType       string
			estimatedStartStr *string
		)

		// When EstimateStart is enabled and time filters are present, we have an extra column
		if filters.EstimateStart && (filters.StartTime != nil || filters.EndTime != nil) {
			err := rows.Scan(&id, &dateCreated, &playbackDuration, &title, &preview, &recordingDuration, &sessionType, &estimatedStartStr)
			if err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
		} else {
			err := rows.Scan(&id, &dateCreated, &playbackDuration, &title, &preview, &recordingDuration, &sessionType)
			if err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
		}

		originalDateTime := dateCreated.Local()
		startDateTime := originalDateTime

		// Determine duration: prefer recording duration, fallback to playback duration
		duration := 0.0
		if recordingDuration != nil {
			duration = *recordingDuration
		} else if playbackDuration != nil {
			duration = *playbackDuration
		}

		// Estimate start time if requested
		if filters.EstimateStart && duration > 0 {
			startDateTime = estimateStartTime(originalDateTime, duration)
		}

		// Truncate preview if too long (using UTF-8 aware truncation)
		preview = utils.TruncateString(preview, 30)

		meetings = append(meetings, MeetingInfo{
			SessionID:         utils.EncodeSessionID(id),
			DateCreated:       originalDateTime,
			DateStarted:       startDateTime,
			Duration:          duration,
			Title:             title,
			Preview:           preview,
			Type:              sessionType,
			UseEstimatedStart: filters.EstimateStart,
		})
	}

	log("Query returned %d rows", len(meetings))

	return meetings, nil
}

// GetSessionByID retrieves session information by session ID
func (db *DB) GetSessionByID(sessionID string, estimateStart bool) (*MeetingInfo, error) {
	id, err := utils.DecodeSessionID(sessionID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT
			s.id,
			s.dateCreated,
			s.playbackDuration,
			%[1]s,
			COALESCE(s.textPreview, ''),
			%[2]s,
			%[3]s
		FROM session s
		%[4]s
		WHERE s.id = ?
	`, sessionTitleExpr, sessionDurationExpr, sessionTypeExpr, sessionSourceJoins)

	var (
		dbID              []byte
		dateCreated       time.Time
		playbackDuration  *float64
		title             string
		preview           string
		recordingDuration *float64
		sessionType       string
	)

	err = db.conn.QueryRow(query, id).Scan(&dbID, &dateCreated, &playbackDuration, &title, &preview, &recordingDuration, &sessionType)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	originalDateTime := dateCreated.Local()
	startDateTime := originalDateTime

	duration := 0.0
	if recordingDuration != nil {
		duration = *recordingDuration
	} else if playbackDuration != nil {
		duration = *playbackDuration
	}

	if estimateStart && duration > 0 {
		startDateTime = estimateStartTime(originalDateTime, duration)
	}

	return &MeetingInfo{
		SessionID:         sessionID,
		DateCreated:       originalDateTime,
		DateStarted:       startDateTime,
		Duration:          duration,
		Title:             title,
		Preview:           preview,
		Type:              sessionType,
		UseEstimatedStart: estimateStart,
	}, nil
}

// GetTranscriptLines retrieves transcript lines for a session
func (db *DB) GetTranscriptLines(sessionID string, estimateStart bool) ([]TranscriptExport, *MeetingInfo, error) {
	id, err := utils.DecodeSessionID(sessionID)
	if err != nil {
		return nil, nil, err
	}

	// First get session info
	info, err := db.GetSessionByID(sessionID, estimateStart)
	if err != nil {
		return nil, nil, err
	}

	query := `
		SELECT
			tl.text,
			tl.start,
			COALESCE(sp.name, 'Unknown Speaker')
		FROM transcriptline tl
		LEFT JOIN speaker sp ON tl.speakerID = sp.id
		WHERE tl.sessionId = ?
		ORDER BY tl.start ASC
	`

	rows, err := db.conn.Query(query, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transcript lines: %w", err)
	}
	defer rows.Close()

	var transcripts []TranscriptExport
	for rows.Next() {
		var (
			text        string
			startMs     int64
			speakerName string
		)

		err := rows.Scan(&text, &startMs, &speakerName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan transcript line: %w", err)
		}

		export := TranscriptExport{
			Speaker: speakerName,
			Text:    text,
		}

		// Calculate actual speak time if needed
		if estimateStart {
			speakTime := info.DateStarted.Add(time.Duration(startMs) * time.Millisecond)
			export.SpeakedAt = &speakTime
		}

		transcripts = append(transcripts, export)
	}

	return transcripts, info, nil
}

// estimateStartTime estimates meeting start time by subtracting duration from dateCreated
// and rounding to nearest 30 minutes
func estimateStartTime(dateCreated time.Time, durationSeconds float64) time.Time {
	// Subtract duration
	estimated := dateCreated.Add(-time.Duration(durationSeconds) * time.Second)

	// Round to nearest 30 minutes
	minute := estimated.Minute()
	roundedMinute := 0
	if minute >= 15 && minute < 45 {
		roundedMinute = 30
	} else if minute >= 45 {
		roundedMinute = 0
		estimated = estimated.Add(time.Hour)
	}

	// Set to rounded time
	return time.Date(
		estimated.Year(),
		estimated.Month(),
		estimated.Day(),
		estimated.Hour(),
		roundedMinute,
		0,
		0,
		estimated.Location(),
	)
}
