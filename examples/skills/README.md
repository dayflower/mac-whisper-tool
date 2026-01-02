# MacWhisper Agent Skills

Example Agent Skills for accessing MacWhisper meeting transcriptions through Claude Code.

## Overview

This directory contains two Agent Skills that provide access to MacWhisper's meeting transcription database:

1. **macwhisper-search** - Search for meetings by keywords, title, and date range
2. **macwhisper-read** - Retrieve full transcription content by session ID

These skills enable Claude to help you find and read meeting transcriptions stored in MacWhisper.

## Quick Start

### Installation

Copy the skills to your Claude skills directory:

```bash
# Copy both skills
cp -r macwhisper-search ~/.claude/skills/
cp -r macwhisper-read ~/.claude/skills/

# Or copy them to a project-specific location
cp -r macwhisper-search /path/to/project/.claude/skills/
cp -r macwhisper-read /path/to/project/.claude/skills/
```

After copying, restart Claude Code to discover the new skills.

### Prerequisites

Before using these skills, ensure:

1. **mac-whisper-tool is installed**: The CLI tool must be available in your PATH
   ```bash
   # Install via Homebrew
   brew install dayflower/tap/mac-whisper-tool

   # Or build from source
   cd /path/to/mac-whisper-tool
   go build -o ~/bin/mac-whisper-tool
   ```

2. **MacWhisper database exists**: The database should be at the default location
   - Default path: `~/Library/Application Support/MacWhisper/Database/main.sqlite`
   - MacWhisper must be installed and have recorded some meetings

3. **Verify installation**:
   ```bash
   # Check if CLI is available
   which mac-whisper-tool

   # Test with a simple list command
   mac-whisper-tool list -n 5
   ```

## Skills Overview

### macwhisper-search

**Purpose**: Search for meeting transcriptions in the MacWhisper database.

**When to use**:
- "Find meetings about [topic]"
- "Search for meetings from last week"
- "Which meetings discussed [keyword]?"

**Capabilities**:
- Content keyword search (full-text in transcriptions)
- Title keyword search
- Date range filtering
- Result limiting (default: 10 meetings)
- Table or JSON output

**Example requests**:
- "Search for meetings about Q4 planning"
- "Find team standup meetings from December"
- "Show all Zoom meetings"

**Output**: Returns a list of matching meetings with session IDs, start times, duration, titles, and previews.

**See**: [macwhisper-search/SKILL.md](macwhisper-search/SKILL.md)

### macwhisper-read

**Purpose**: Retrieve the full transcription of a specific meeting.

**When to use**:
- "Read the content of meeting [session-id]"
- "Show me the full transcription"
- After searching, to view complete meeting content

**Capabilities**:
- Fetch full transcription by session ID
- Includes metadata (title, dates, duration)
- Includes timestamps for each line
- Speaker identification

**Example requests**:
- "Read meeting ABC123"
- "Show me the full transcription of that meeting"
- "What was discussed in meeting XYZ789?"

**Output**: Returns extended Markdown format with metadata, timestamps, speakers, and full transcript text.

**See**: [macwhisper-read/SKILL.md](macwhisper-read/SKILL.md)

## Typical Workflow

### 1. Search for Meetings

Use **macwhisper-search** to find meetings:

```
User: "Find meetings about budget planning"

Claude uses: macwhisper-search skill
→ Searches with keywords "budget" and "planning"
→ Returns matching meetings with session IDs
```

### 2. Read Specific Meeting

Use **macwhisper-read** to retrieve full content:

```
User: "Read the first meeting"

Claude uses: macwhisper-read skill
→ Uses session ID from search results
→ Returns full transcription with timestamps
```

### 3. Combined Workflow Example

```
User: "Find and show me Q4 planning meetings"

Step 1: Claude searches for meetings
  → macwhisper-search with "Q4" and "planning" keywords

Step 2: Claude presents results
  → Shows list of matching meetings

Step 3: User asks to read one
  → "Show me the first one"

Step 4: Claude reads the transcription
  → macwhisper-read with the session ID
  → Displays full meeting content
```

## Configuration

### Custom Database Path

If your MacWhisper database is not at the default location, you can configure it in multiple ways:

**Recommended: Configuration File** (best for Agent Skills):

Create `~/.config/MacWhisperTool/config.json`:

```json
{
  "database": {
    "path": "/path/to/custom/main.sqlite"
  }
}
```

**Alternative: Environment Variable**:

```bash
export MAC_WHISPER_DB="/path/to/custom/main.sqlite"
```

For permanent configuration, add to your shell profile (`~/.zshrc` or `~/.bashrc`).

**Priority order**:
1. `--db` flag (explicit override)
2. `MAC_WHISPER_DB` environment variable
3. `~/.config/MacWhisperTool/config.json`
4. `~/Library/Application Support/MacWhisperTool/config.json`
5. Default path (`~/Library/Application Support/MacWhisper/Database/main.sqlite`)

### Default Settings

Both skills use these defaults:

- **Start time estimation**: Enabled (`--estimate-start`)
  - Estimates meeting start by subtracting duration from end time
  - Rounds to nearest 30 minutes

- **Search result limit**: 10 meetings (matching MCP server default)

- **Output format**:
  - Search: Table format (human-readable)
  - Read: Extended Markdown (with metadata and timestamps)

## Session IDs

Session IDs are base64-encoded 128-bit identifiers without padding.

**Format**: `aBcDeFgH1234567890XyZ`

**Characteristics**:
- Case-sensitive
- No special characters (alphanumeric only)
- Fixed encoding (base64 without padding)

**How to obtain**:
1. From search results (first column in table output)
2. From JSON output (`sessionId` field)
3. Displayed in Claude's responses when using macwhisper-search

## Date and Time Formats

### Input Formats

When searching by date:

- Date only: `2025-12-15` (treated as 00:00:00)
- Date and time: `2025-12-15T14:30:00`
- With milliseconds: `2025-12-15T14:30:00.000`

### Output Format

All timestamps use ISO 8601 without timezone:

- `2025-12-15T14:00:00.000`

## Examples

### Search Examples

```bash
# Find meetings about specific topic
"Search for meetings about quarterly review"

# Find meetings from date range
"Show meetings from last week"

# Find by title
"Find all standup meetings"
```

### Read Examples

```bash
# Read specific meeting
"Read meeting ABC123XYZ"

# Read from search results
"Show me the first meeting from those results"

# Read and analyze
"Read the Q4 planning meeting and summarize the key points"
```

## Troubleshooting

### Skills Not Discovered

**Problem**: Claude doesn't recognize the skills

**Solution**:
1. Verify skills are in `~/.claude/skills/` or `.claude/skills/`
2. Check SKILL.md files have correct frontmatter (YAML header)
3. Restart Claude Code
4. Try explicit invocation: "Use macwhisper-search to find meetings about X"

### CLI Not Found

**Problem**: "mac-whisper-tool: command not found"

**Solution**:
1. Verify installation: `which mac-whisper-tool`
2. Add to PATH: `export PATH="$HOME/bin:$PATH"`
3. Reinstall if necessary

### Database Not Found

**Problem**: "failed to open database" error

**Solution**:
1. Verify MacWhisper is installed
2. Check database exists: `ls -la ~/Library/Application\ Support/MacWhisper/Database/main.sqlite`
3. Specify custom path if database is elsewhere

### No Search Results

**Problem**: Search returns no results

**Solution**:
- Try broader keywords
- Remove date filters
- Verify meetings exist in database: `mac-whisper-tool list`

## Additional Resources

- **Main Project**: [mac-whisper-tool](https://github.com/dayflower/mac-whisper-tool)
- **CLI Documentation**: See main README.md
- **MCP Server**: See sketch/MCP_PLAN.md for MCP server details
- **Skills Documentation**:
  - [macwhisper-search/SKILL.md](macwhisper-search/SKILL.md)
  - [macwhisper-search/EXAMPLES.md](macwhisper-search/EXAMPLES.md)
  - [macwhisper-read/SKILL.md](macwhisper-read/SKILL.md)
  - [macwhisper-read/EXAMPLES.md](macwhisper-read/EXAMPLES.md)

## Contributing

These skills are provided as examples. Feel free to:

- Customize for your workflow
- Add additional filters or options
- Submit improvements via pull request

## License

Same license as the main mac-whisper-tool project.
