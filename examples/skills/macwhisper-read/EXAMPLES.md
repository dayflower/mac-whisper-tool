# MacWhisper Read - Usage Examples

This document provides common usage examples for the macwhisper-read skill.

## Basic Use Cases

### 1. Read a Specific Meeting

**Scenario**: You have a session ID from a search result

**Command**:
```bash
mac-whisper-tool export -x --estimate-start "ABC123XYZ"
```

**Output**:
```markdown
# Zoom Meeting

- **Date Started**: 2025-12-15T14:00:00.000
- **Date Created**: 2025-12-15T15:45:30.000
- **Duration**: 1h 45m 30s

- 2025-12-15T14:00:00.000 **Speaker 1**: Good afternoon, everyone.
- 2025-12-15T14:00:15.234 **Speaker 2**: Hi, thanks for joining.
- 2025-12-15T14:01:03.456 **Speaker 1**: Let's start with the agenda...
```

### 2. Complete Search-to-Read Workflow

**Step 1**: Search for meetings about "Q4 planning"

```bash
mac-whisper-tool search -k "Q4" -k "planning" --estimate-start -n 10
```

**Output**:
```
Session ID         Start Time            Duration  Title                Preview
xY9zA1bC2d        2025-12-15 14:00      1h 45m    Q4 Planning Call     We need to discuss...
aB3cD4eF5g        2025-12-10 10:00      45m 20s   Q4 Budget Planning   Let's review the...
```

**Step 2**: Read the full transcription of the first meeting

```bash
mac-whisper-tool export -x --estimate-start "xY9zA1bC2d"
```

**Result**: Full transcription with all details

### 3. Obtaining Session IDs from JSON Search

**Step 1**: Search with JSON output

```bash
mac-whisper-tool search -k "design review" --estimate-start -f json
```

**Output**:
```json
[
  {
    "sessionId": "pQ7rS8tU9v",
    "dateCreated": "2025-12-15T15:45:30.000",
    "dateStarted": "2025-12-15T14:00:00.000",
    "duration": 2730.0,
    "title": "Design Review Meeting",
    "preview": "Today we're reviewing..."
  }
]
```

**Step 2**: Extract session ID and read

```bash
mac-whisper-tool export -x --estimate-start "pQ7rS8tU9v"
```

## Output Examples

### Example 1: Short Meeting (15 minutes)

```markdown
# Daily Standup

- **Date Started**: 2025-12-15T09:00:00.000
- **Date Created**: 2025-12-15T09:15:12.000
- **Duration**: 15m 12s

- 2025-12-15T09:00:00.000 **Speaker 1**: Good morning, team. Let's do a quick standup.
- 2025-12-15T09:00:23.456 **Speaker 2**: I finished the authentication module yesterday.
- 2025-12-15T09:01:45.678 **Speaker 3**: I'm working on the database migration today.
- 2025-12-15T09:03:12.890 **Speaker 1**: Great. Any blockers?
- 2025-12-15T09:03:30.123 **Speaker 2**: No blockers on my end.
```

### Example 2: Long Meeting (2+ hours)

```markdown
# All-Hands Q4 Review

- **Date Started**: 2025-12-20T14:00:00.000
- **Date Created**: 2025-12-20T16:25:45.000
- **Duration**: 2h 25m 45s

- 2025-12-20T14:00:00.000 **Speaker 1**: Welcome everyone to our Q4 all-hands.
- 2025-12-20T14:00:30.123 **Speaker 1**: We have a lot to cover today...
- 2025-12-20T14:05:15.456 **Speaker 2**: Let me share the Q4 metrics.
...
- 2025-12-20T16:24:30.789 **Speaker 1**: Thanks everyone for a productive session.
- 2025-12-20T16:25:00.012 **Speaker 3**: Thank you!
```

## Integration Patterns

### Pattern 1: Find and Read Most Recent Meeting

```bash
# Step 1: Get the latest meeting
mac-whisper-tool search -s $(date -v-1d +%Y-%m-%d) --estimate-start -n 1 -f json

# Step 2: Extract session ID (manual or scripted)
# Step 3: Read it
mac-whisper-tool export -x --estimate-start "sessionID"
```

### Pattern 2: Read All Meetings from a Specific Date

```bash
# Step 1: Search for meetings on a specific date
mac-whisper-tool search -s 2025-12-15 -e 2025-12-15 --estimate-start -f json

# Step 2: For each session ID in the results, read the transcription
# (This would typically be done programmatically)
```

## Troubleshooting

### Error: "Session not found"

**Problem**: The session ID doesn't exist in the database

**Possible causes**:
- Typo in session ID
- Session ID from a different database
- Meeting was deleted

**Solution**:
- Verify the session ID from search results
- Ensure you're using the correct database

**Example**:
```bash
# Wrong (non-existent session ID)
mac-whisper-tool export -x --estimate-start "INVALID123"
# Error: failed to get transcript: Session not found

# Correct (valid session ID from search)
mac-whisper-tool export -x --estimate-start "xY9zA1bC2d"
# Success: Returns transcription
```

### Error: "Session ID required"

**Problem**: No session ID provided

**Cause**: Missing position argument

**Solution**: Add session ID as the last argument

```bash
# Wrong (no session ID)
mac-whisper-tool export -x --estimate-start
# Error: Session ID required

# Correct
mac-whisper-tool export -x --estimate-start "ABC123"
# Success
```

### Database Not Found Error

**Problem**: "failed to open database" error

**Cause**: Database file not at default location

**Solution**: Ensure MacWhisper is installed, or specify custom database path:

```bash
mac-whisper-tool export -x --estimate-start --db /path/to/main.sqlite "sessionID"
```

## Tips

1. **Always use search first**: Use `macwhisper-search` to find session IDs before reading

2. **Session IDs are case-sensitive**: Copy them exactly as shown in search results

3. **Timestamps show when content was spoken**: Use timestamps to find specific moments in meetings

4. **Duration includes all spoken time**: The duration reflects the actual meeting length

5. **Multiple speakers are identified**: Each speaker gets a unique identifier (Speaker 1, Speaker 2, etc.)

## Advanced Use Cases

### Finding Specific Discussion Points

After reading a transcription, you can search within the output for specific topics:

```bash
# Read transcription and search for "action items"
mac-whisper-tool export -x --estimate-start "ABC123" | grep -i "action item"
```

### Extracting Just the Metadata

```bash
# Read transcription and extract just the metadata section
mac-whisper-tool export -x --estimate-start "ABC123" | head -n 5
```

### Comparing Two Meetings

```bash
# Read two meetings and compare
mac-whisper-tool export -x --estimate-start "SESSION1" > meeting1.md
mac-whisper-tool export -x --estimate-start "SESSION2" > meeting2.md
# Then use diff or other tools to compare
```

## Understanding the Output

### Speaker Identification

- **Speaker 1, Speaker 2, etc.**: MacWhisper automatically identifies different speakers
- **Consistent across meeting**: Same speaker keeps the same identifier throughout
- **No name attribution**: Speakers are numbered, not named

### Timestamp Precision

- **Millisecond precision**: Timestamps include milliseconds (e.g., `.456`)
- **Relative to start**: All timestamps are calculated from the estimated start time
- **Rounded start time**: Start time is rounded to nearest 30 minutes

### Duration Format

Duration is displayed in human-readable format:

- Short: `15m 12s` (15 minutes, 12 seconds)
- Medium: `1h 15m 30s` (1 hour, 15 minutes, 30 seconds)
- Long: `2h 25m 45s` (2 hours, 25 minutes, 45 seconds)
