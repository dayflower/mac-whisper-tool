---
name: macwhisper-read
description: Read MacWhisper meeting transcription by session ID. Returns the full transcript in Markdown format with metadata and timestamps. Use when you need to retrieve or view the content of a specific meeting.
allowed-tools: Bash(mac-whisper-tool export:*)
---

# MacWhisper Read Skill

Retrieve the full transcription of a meeting by its session ID.

## Prerequisites

- `mac-whisper-tool` must be installed and available in PATH
- MacWhisper database at default location: `~/Library/Application Support/MacWhisper/Database/main.sqlite`
- Session ID (obtained from `macwhisper-search` skill or other sources)

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

## What This Skill Does

This skill reads a meeting transcription from the MacWhisper database and returns it in an extended Markdown format that includes:

- Meeting title
- Metadata (Date Started, Date Created, Duration)
- Full transcription with timestamps for each line
- Speaker identification

## Usage

### Read a Meeting by Session ID

```bash
mac-whisper-tool export -x --estimate-start "sessionID"
```

Replace `sessionID` with the actual session ID (base64-encoded identifier).

## Parameters

| Parameter | CLI Flag | Description |
|-----------|----------|-------------|
| Session ID | Position argument (required) | The base64-encoded session identifier |
| Extended format | `-x` | Include metadata and timestamps (always used) |
| Estimate start | `--estimate-start` | Estimate meeting start time (always used) |
| Output | stdout | Output destination (stdout by default) |

## Default Behavior

The skill uses these defaults:

- **Extended format**: Always enabled (`-x`) for full metadata and timestamps
- **Start time estimation**: Always enabled (`--estimate-start`) for accurate timestamps
- **Output format**: Markdown
- **Output destination**: stdout (displayed to user or captured by Claude)

## Session IDs

Session IDs are base64-encoded 128-bit identifiers without padding.

Example: `aBcDeFgH1234567890XyZ`

### How to Get Session IDs

1. **From search results**: Use the `macwhisper-search` skill to find meetings
2. **From table output**: The first column shows session IDs
3. **From JSON output**: The `sessionId` field contains the ID

## Output Format

The skill returns transcriptions in extended Markdown format:

```markdown
# Meeting Title

- **Date Started**: 2025-07-23T10:00:00.000
- **Date Created**: 2025-07-23T11:59:59.000
- **Duration**: 38m 52s

- 2025-07-23T10:00:00.000 **Speaker 1**: Hello everyone, let's start the meeting.
- 2025-07-23T10:01:23.456 **Speaker 2**: Thanks for joining.
- 2025-07-23T10:02:45.789 **Speaker 1**: Today's agenda includes...
```

### Metadata Section

The metadata section includes:

- **Date Started**: Estimated meeting start time (rounded to nearest 30 minutes)
- **Date Created**: When the transcription was finalized
- **Duration**: Total meeting duration in human-readable format (e.g., "38m 52s", "1h 15m 30s")

### Transcription Lines

Each transcription line includes:

- **Timestamp**: When the line was spoken (ISO 8601 format)
- **Speaker**: Speaker identifier (e.g., "Speaker 1", "Speaker 2")
- **Text**: The actual spoken content

## Error Handling

- **Invalid session ID**: Returns "Session not found" error
- **Session ID not provided**: Returns error - session ID is required
- **Database error**: Passes through CLI error message

## Date/Time Formats

### Output Format

All timestamps use ISO 8601 format without timezone:

- Date and time: `2025-07-23T10:00:00.000`

### Start Time Estimation

When `--estimate-start` is enabled (always in this skill):

1. Calculates start time by subtracting duration from `dateCreated`
2. Rounds to nearest 30 minutes
3. Uses this estimated start time for all transcription line timestamps

Example:
- Meeting ended: 2025-07-23T11:59:59.000
- Duration: 1h 45m 30s
- Calculated start: 2025-07-23T10:14:29.000
- Rounded start: 2025-07-23T10:00:00.000

## Typical Workflow

### 1. Search for Meetings

First, use `macwhisper-search` to find meetings:

```bash
mac-whisper-tool search -k "design review" --estimate-start -n 10
```

### 2. Note the Session ID

From the search results, identify the session ID:

```
Session ID         Start Time            Duration  Title
xY9zA1bC2d        2025-12-15 14:00      45m 30s   Design Review
```

### 3. Read the Full Transcription

Use this skill to retrieve the full content:

```bash
mac-whisper-tool export -x --estimate-start "xY9zA1bC2d"
```

### 4. View the Content

The full transcription with timestamps and metadata is displayed.

## Examples

See [EXAMPLES.md](EXAMPLES.md) for detailed usage examples and common use cases.

## Related Skills

- **macwhisper-search**: Search for meetings and obtain session IDs

## Technical Details

### Command Explanation

The underlying CLI command:

```bash
mac-whisper-tool export -x --estimate-start "sessionID"
```

- `export`: Uses the export command
- `-x`: Enables extended format (metadata + timestamps)
- `--estimate-start`: Enables start time estimation
- `"sessionID"`: The session identifier (position argument)

### Output Destination

- **Default**: stdout (printed to console)
- **No file output**: This skill doesn't write to files - it streams content to stdout for immediate use

This design allows Claude to directly read and process the transcription content.
