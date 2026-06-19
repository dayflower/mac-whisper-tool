# mac-whisper-tool

A CLI tool to export meeting transcriptions from MacWhisper's database.

## Project Overview

This tool provides command-line access to MacWhisper's SQLite database, enabling users to list and export meeting transcriptions in various formats (Markdown, JSON).

## Technical Stack

- **Language**: Go
- **Database**: SQLite3 (MacWhisper's database)
- **CLI Framework**: Use `cobra` for subcommand structure
- **Database Access**: Use `github.com/ncruces/go-sqlite3` (pure Go, no CGO required)
- **Table Output**: Use `github.com/olekukonko/tablewriter` for table formatting
- **Date/Time**: Standard library `time` package with JST timezone support

## Project Structure

```
mac-whisper-tool/
├── cmd/
│   ├── root.go           # Root command and global flags
│   ├── list.go           # List subcommand
│   ├── export.go         # Export subcommand
│   ├── mcp.go            # MCP subcommand (Phase 2)
│   └── interactive.go    # Interactive subcommand (Phase 3)
├── internal/
│   ├── db/
│   │   ├── db.go         # Database connection and queries
│   │   └── models.go     # Data models
│   ├── exporter/
│   │   ├── markdown.go   # Markdown export
│   │   └── json.go       # JSON export
│   ├── formatter/
│   │   └── table.go      # Table formatting for list
│   ├── mcp/              # MCP server implementation (Phase 2)
│   │   ├── server.go     # MCP server main
│   │   ├── tools.go      # Tool handler (search_meetings)
│   │   ├── resources.go  # Resource handlers
│   │   ├── search.go     # Search logic
│   │   └── uri.go        # URI parsing
│   └── utils/
│       ├── datetime.go   # Date/time utilities (estimation, parsing, formatting)
│       └── sessionid.go  # Session ID encoding (binary to base64)
├── go.mod
├── go.sum
├── main.go
├── README.md
└── CLAUDE.md
```

## Key Implementation Details

### Session ID Handling

- MacWhisper stores 128-bit binary session IDs
- Convert to base64 without padding for CLI usage
- Use `base64.RawStdEncoding` (standard base64 without padding)

### Date/Time Handling

- Database stores timestamps as standard SQLite DATETIME (not Core Data format)
- Convert to local timezone for display
- Support two modes:
  1. Default: Use `dateCreated` from database
  2. With `--estimate-start`: Estimate meeting start time by:
     - Subtracting `duration` from `dateCreated`
     - Rounding to nearest 30 minutes
- Output format: ISO 8601 without timezone (e.g., `2025-07-23T11:22:33.000`)
- Input parsing: Accept `YYYY-MM-DD`, `YYYY-MM-DDTHH:MM:SS`, or `YYYY-MM-DDTHH:MM:SS.sss`

### Database Schema

- `ZRECORDEDMEETING` table for meeting metadata
- `ZSESSION` table for session data
- `ZTRANSCRIPTLINE` table for transcript lines

### Export Formats and Content Modes

**Format Options** (`--format`, `-f`):

- `markdown` (or `md`): Markdown format (default)
- `json`: JSON format

**Content Options** (`--extend`, `-x`):

- Default: Standard content (MacWhisper compatible)
- `-x`: Extended content (includes timestamps and metadata)

**Start Time Estimation** (`--estimate-start`):

- Estimates meeting start time (internal calculation)
- Doesn't affect output format by itself
- When combined with `-x`, timestamps use estimated start time

1. **Markdown**

   - Standard: `- **Speaker 1**: text`
   - Extended (`-x`): Includes title header, metadata (Date Started, Date Created, Duration), and timestamps for each line

2. **JSON**

   - Standard: Array of `{speaker, text}` objects (MacWhisper compatible)
   - Extended (`-f json -x`): Full metadata with `{title, dateStarted, dateCreated, duration, transcripts[]}` where each transcript includes `speakedAt`

3. **Table** (list command only)
   - Columns: Session ID, Start Time, Duration, Type, Title, Preview
   - `list`/`search` return all session source types (recorded-meeting, system-audio, voice-memo, podcast, youtube, download, imported); `--type` filters to one

### File Naming Convention

Auto-generated filenames follow the pattern: `{datetime} {title}.{ext}`

- Example: `2025-07-23T12:34:56.000 Zoom Meeting.md`

### Error Handling

- Non-existent session ID: stderr message, exit code 1
- Missing DB file: stderr message, exit code 1
- Missing output directory: Auto-create (fail with exit code 1 on error)
- Batch export without `--output-dir`: stderr message, exit code 1

## Development Guidelines

### Phase 1 (Initial Implementation)

- Implement `list` subcommand
- Implement `export` subcommand (single and batch modes)
- Support all specified output formats
- Support date/time filtering and estimation

### Phase 2 (Completed)

- Implement MCP server subcommand ✅
- Support `search_meetings` tool for content/title/date search ✅
- Expose sessions as Resources with `macwhisper://localhost/session/{id}` URI scheme ✅
- stdio transport only for local MCP client integration ✅
- Legacy mode support for older MCP clients ✅
  - `--legacy` flag to enable legacy mode
  - `get_meeting` tool for retrieving transcriptions (legacy mode only)
  - JSON-formatted search results (both modes)
  - Resources disabled in legacy mode

### Phase 3 (Future)

- Implement `interactive` subcommand with TUI
- Consider using `bubbletea` or `tview` for TUI

### Phase 4 (Future)

- Support multiple selection in interactive mode

## Testing Strategy

- Unit tests for date/time utilities (estimation, rounding, parsing)
- Unit tests for session ID encoding/decoding
- Integration tests with a test SQLite database
- Test edge cases: empty results, invalid dates, missing files

## Dependencies

```go
require (
    github.com/spf13/cobra v1.8.0
    github.com/ncruces/go-sqlite3 v0.30.4
    github.com/olekukonko/tablewriter v0.0.5
    // Phase 2: MCP server support
    github.com/modelcontextprotocol/go-sdk v0.x.x  // TBD
)
```

**Note**: This project uses `github.com/ncruces/go-sqlite3`, a pure Go SQLite driver that doesn't require CGO. This allows for simpler cross-compilation and faster CI builds.

## Default Values

- Default DB path: `~/Library/Application Support/MacWhisper/Database/main.sqlite`
- Default list limit: 20 items
- Default export format: Markdown
- Default output: stdout (unless batch mode or directory specified)

## Verbose Mode

- Use `--verbose` or `-v` flag
- Output progress messages to stderr
- Always output errors to stderr (even in non-verbose mode)
