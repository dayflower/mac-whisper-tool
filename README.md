# mac-whisper-tool

A CLI tool to export meeting transcriptions from [MacWhisper](https://goodsnooze.gumroad.com/l/macwhisper)'s database.

## Overview

This tool provides command-line access to MacWhisper's SQLite database, enabling users to list and export meeting transcriptions in various formats (Markdown, JSON).

## Features

- List meetings with filtering by date range
- Export transcriptions in multiple formats:
  - Markdown (standard or extended with metadata)
  - JSON (standard MacWhisper compatible or extended with metadata)
- Single session export or batch export
- Meeting start time estimation

## Installation

### Pre-built Binaries

Download the latest release for your macOS architecture from the [Releases](https://github.com/dayflower/mac-whisper-tool/releases) page:

- **Intel Mac (x86_64)**: Download `mac-whisper-tool_*_darwin_x86_64.tar.gz`
- **Apple Silicon (ARM64)**: Download `mac-whisper-tool_*_darwin_arm64.tar.gz`

Extract and install:

```bash
# Extract the archive
tar -xzf mac-whisper-tool_*_darwin_*.tar.gz

# Move to a directory in your PATH (optional)
sudo mv mac-whisper-tool /usr/local/bin/
```

### From Source

```bash
go build -o mac-whisper-tool
```

The binary will be created in the current directory.

## Usage

### List Meetings

Display a list of meetings from the database:

```bash
# Show the most recent 20 meetings (default)
mac-whisper-tool list

# Show all meetings in JSON format
mac-whisper-tool list -n -1 -f json

# Filter by date range
mac-whisper-tool list -s 2025-12-01 -e 2025-12-31

# Estimate meeting start times (rounds to nearest 30 minutes)
mac-whisper-tool list --estimate-start
```

**Options:**

- `-d, --db <path>` - Database file path (default: `~/Library/Application Support/MacWhisper/Database/main.sqlite`)
- `-s, --start <datetime>` - Filter by start date (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)
- `-e, --end <datetime>` - Filter by end date
- `-n, --limit <n>` - Maximum number of meetings (default: 20, negative for all)
- `-f, --format <format>` - Output format: `table` or `json` (default: table)
- `--estimate-start` - Estimate meeting start time from dateCreated and duration
- `-v, --verbose` - Enable verbose output to stderr

### Export Transcriptions

Export a single meeting transcription:

```bash
# Export to stdout in Markdown format (default)
mac-whisper-tool export <session-id>

# Export to a specific file
mac-whisper-tool export -o output.md <session-id>

# Export to a directory with auto-generated filename
mac-whisper-tool export -c ./exports <session-id>

# Export in JSON format (MacWhisper compatible)
mac-whisper-tool export -f json <session-id>

# Export in extended Markdown format with metadata and timestamps
mac-whisper-tool export -x <session-id>

# Export in extended JSON format with metadata and timestamps
mac-whisper-tool export -f json -x <session-id>

# Estimate start time (internal calculation only, doesn't affect output)
mac-whisper-tool export --estimate-start <session-id>

# Estimate start time with extended content
mac-whisper-tool export --estimate-start -x <session-id>
```

**Batch Export:**

Export multiple meetings at once:

```bash
# Export the latest 10 meetings
mac-whisper-tool export -c ./exports -n 10

# Export meetings from a specific date range
mac-whisper-tool export -c ./exports -s 2025-12-01 -e 2025-12-31

# Export all meetings
mac-whisper-tool export -c ./exports -n -1
```

**Options:**

- `-d, --db <path>` - Database file path
- `-f, --format <format>` - Output format: `markdown` (or `md`), `json` (default: markdown)
- `-o, --output <file>` - Output file path (single session only)
- `-c, --output-dir <dir>` - Output directory (filename auto-generated)
- `-x, --extend` - Output extended content (timestamps, metadata)
- `--estimate-start` - Estimate meeting start time from dateCreated and duration
- `-v, --verbose` - Enable verbose output to stderr

**Batch export options:**

- `-s, --start <datetime>` - Filter by start date
- `-e, --end <datetime>` - Filter by end date
- `-n, --limit <n>` - Maximum number of meetings to export (negative for all)

## Output Formats

### Format vs Content

- **Format** (`--format`, `-f`): Specifies the output format

  - `markdown` (or `md`): Markdown format (default)
  - `json`: JSON format

- **Content** (`--extend`, `-x`): Specifies whether to include extended content

  - Default: Standard content (MacWhisper compatible)
  - `-x`: Extended content (includes timestamps and metadata)

- **Start Time Estimation** (`--estimate-start`): Estimates meeting start time
  - Internal calculation only, doesn't affect output format
  - When combined with `-x`, timestamps are based on estimated start time

### Markdown

**Standard format (default):**

```markdown
- **Speaker 1**: foo bar baz
- **Speaker 2**: hoge fuga
```

**Extended format (`-x`):**

```markdown
# Zoom Meeting

- **Date Started**: 2025-07-23T10:00:00.000
- **Date Created**: 2025-07-23T11:59:59.000
- **Duration**: 38m 52s

- 2025-07-23T10:00:00.000 **Speaker 1**: foo bar baz
- 2025-07-23T10:20:30.400 **Speaker 2**: hoge fuga
```

### JSON

**Standard format (MacWhisper compatible):**

```json
[
  {
    "speaker": "Speaker 1",
    "text": "foo bar baz"
  },
  {
    "speaker": "Speaker 2",
    "text": "hoge fuga"
  }
]
```

**Extended format (`-f json -x`):**

```json
{
  "title": "Zoom Meeting",
  "dateStarted": "2025-07-23T10:00:00.000",
  "dateCreated": "2025-07-23T11:59:59.000",
  "duration": 2332.4,
  "transcripts": [
    {
      "speaker": "Speaker 1",
      "text": "foo bar baz",
      "speakedAt": "2025-07-23T10:00:00.000"
    },
    {
      "speaker": "Speaker 2",
      "text": "hoge fuga",
      "speakedAt": "2025-07-23T10:20:30.400"
    }
  ]
}
```

## Auto-generated Filenames

When using `-c, --output-dir`, filenames are automatically generated in the format:

```
{datetime} {title}.{ext}
```

Example: `2025-07-23T12:34:56.000 Zoom Meeting.md`

Invalid filename characters are replaced with underscores.

## Date/Time Handling

### Input Formats

The tool accepts the following datetime formats:

- Date only: `2025-07-23` (treated as `2025-07-23T00:00:00.000`)
- Date and time: `2025-07-23T11:22:33`
- Date and time with milliseconds: `2025-07-23T11:22:33.000`

### Start Time Estimation

When using `--estimate-start`, the tool estimates the meeting start time by:

1. Subtracting the meeting duration from `dateCreated`
2. Rounding to the nearest 30 minutes

For example:

- Calculated time: 11:55 → Rounded to: 12:00
- Calculated time: 13:10 → Rounded to: 13:00

## Database Location

The default database path is:

```
~/Library/Application Support/MacWhisper/Database/main.sqlite
```

You can specify a different path using the `-d, --db` flag.

## Error Handling

- **Non-existent session ID**: Error message to stderr, exit code 1
- **Missing database file**: Error message to stderr, exit code 1
- **Missing output directory**: Automatically created (exit code 1 on failure)
- **Batch export without --output-dir**: Error message to stderr, exit code 1

## Examples

```bash
# List the latest 5 meetings with verbose output
mac-whisper-tool list -n 5 -v

# Export a specific meeting to a file with extended content
mac-whisper-tool export -x -o meeting.md abc123def456

# Batch export all meetings from December 2025 in extended JSON format
mac-whisper-tool export -c ./exports -f json -x -s 2025-12-01 -e 2025-12-31 -n -1

# Export using a custom database path with estimated start time
mac-whisper-tool export -d /path/to/custom.sqlite --estimate-start -x -c ./output sessionID
```

## License

See [LICENSE](./LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
