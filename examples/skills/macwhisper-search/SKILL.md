---
name: macwhisper-search
description: Search MacWhisper meeting transcriptions by keywords, title, and date range. Use when searching for meetings, finding transcripts, or looking for specific content in meeting recordings.
allowed-tools: Bash(mac-whisper-tool search:*), Bash(mac-whisper-tool list:*)
---

# MacWhisper Search Skill

Search for meeting transcriptions in the MacWhisper database by keywords, title, or date range.

## Prerequisites

- `mac-whisper-tool` must be installed and available in PATH
- MacWhisper database at default location: `~/Library/Application Support/MacWhisper/Database/main.sqlite`

## Custom Database Path

If your MacWhisper database is not at the default location, create a configuration file:

**Recommended method** (persistent configuration):

```bash
mkdir -p ~/.config/MacWhisperTool
cat > ~/.config/MacWhisperTool/config.json <<EOF
{
  "database": {
    "path": "/path/to/custom/main.sqlite"
  }
}
EOF
```

**Alternative** (environment variable):

```bash
export MAC_WHISPER_DB="/path/to/custom/main.sqlite"
```

Both methods are automatically recognized by all commands.

## Search Capabilities

This skill provides:

- **Content keyword search**: Search within meeting transcriptions (AND condition for multiple keywords)
- **Title keyword search**: Search within meeting titles (AND condition for multiple keywords)
- **Date range filtering**: Filter by meeting start/end dates
- **Result limiting**: Control the number of results (default: 10)
- **Output formats**: Table format for human reading, JSON for programmatic use

## Usage

### Basic Keyword Search

To search for meetings containing specific keywords in the transcription:

```bash
mac-whisper-tool search -k "keyword" --estimate-start -n 10
```

### Multiple Keywords (AND Condition)

Search for meetings containing ALL specified keywords:

```bash
mac-whisper-tool search -k "zoom" -k "meeting" --estimate-start -n 10
```

### Title Search

Search for meetings by title keywords:

```bash
mac-whisper-tool search -t "standup" --estimate-start -n 10
```

### Combined Search

Combine content keywords, title keywords, and date range:

```bash
mac-whisper-tool search -k "Q4" -t "planning" -s 2025-10-01 -e 2025-12-31 --estimate-start -n 10
```

### Date Range Filtering

Filter meetings by date range without keywords:

```bash
mac-whisper-tool search -s 2025-12-01 -e 2025-12-31 --estimate-start -n 10
```

### JSON Output

Get results in JSON format for programmatic processing:

```bash
mac-whisper-tool search -k "project" --estimate-start -f json
```

### Unlimited Results

Get all matching results (no limit):

```bash
mac-whisper-tool search -k "meeting" --estimate-start -n -1
```

## Parameters

| Parameter | CLI Flag | Description |
|-----------|----------|-------------|
| Content keywords | `-k keyword` | Search in transcription text (repeatable, AND condition) |
| Title keywords | `-t keyword` | Search in meeting title (repeatable, AND condition) |
| Start date | `-s YYYY-MM-DD` | Filter by start date (also accepts `YYYY-MM-DDTHH:MM:SS`) |
| End date | `-e YYYY-MM-DD` | Filter by end date (also accepts `YYYY-MM-DDTHH:MM:SS`) |
| Result limit | `-n 10` | Maximum number of results (use `-n -1` for unlimited) |
| Output format | `-f table` | Output format: `table` (default) or `json` |
| Estimate start | `--estimate-start` | Estimate meeting start time (always used) |

## Default Behavior

The skill uses these defaults:

- **Start time estimation**: Always enabled (`--estimate-start`)
- **Result limit**: 10 meetings (matching MCP server default)
- **Output format**: Table format for easy reading

## Output

### Table Format

The default table output shows:

- Session ID (base64-encoded)
- Start time (estimated)
- Duration
- Title
- Preview (first few words of transcription)

Example:

```
Session ID         Start Time            Duration  Title          Preview
ABC123XYZ          2025-12-15 14:00      45m 30s   Zoom Meeting   This is the preview...
DEF456UVW          2025-12-14 10:00      15m 12s   Team Standup   Another preview...
```

### JSON Format

JSON output includes full meeting metadata:

```json
[
  {
    "sessionId": "ABC123XYZ",
    "dateCreated": "2025-12-15T15:45:30.000",
    "dateStarted": "2025-12-15T14:00:00.000",
    "duration": 2730.0,
    "title": "Zoom Meeting",
    "preview": "This is the preview..."
  }
]
```

## Session IDs

Session IDs are base64-encoded 128-bit identifiers without padding.

Example: `aBcDeFgH1234567890XyZ`

Use these session IDs with the `macwhisper-read` skill to retrieve full transcriptions.

## Error Handling

- **No search criteria**: Returns error - at least one search criterion (keywords, title, or date range) is required
- **Invalid date format**: Returns error with expected format (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)
- **No results found**: Displays "No matching meetings found"
- **Database error**: Passes through CLI error message

## Date/Time Formats

### Input Formats

- Date only: `2025-12-15` (treated as 00:00:00 local time)
- Date and time: `2025-12-15T14:30:00`
- Date and time with milliseconds: `2025-12-15T14:30:00.000`

### Output Format

- ISO 8601 without timezone: `2025-12-15T14:00:00.000`

### Start Time Estimation

When `--estimate-start` is enabled (always in this skill):

1. Calculates start time by subtracting duration from `dateCreated`
2. Rounds to nearest 30 minutes

Example:
- Calculated time: 11:55 → Rounded to: 12:00
- Calculated time: 13:10 → Rounded to: 13:00

## Examples

See [EXAMPLES.md](EXAMPLES.md) for more detailed usage examples and common use cases.

## Related Skills

- **macwhisper-read**: Retrieve full transcription content by session ID
