# MacWhisper Search - Usage Examples

This document provides common usage examples for the macwhisper-search skill.

## Basic Use Cases

### 1. Find Meetings About a Specific Topic

**User request**: "Search for meetings about Q4 planning"

**Command**:
```bash
mac-whisper-tool search -k "Q4" -k "planning" --estimate-start -n 10
```

**What happens**:
- Searches for meetings containing both "Q4" AND "planning" in transcription
- Returns up to 10 results
- Estimates meeting start times

### 2. Find Recent Team Standups

**User request**: "Find team standup meetings from last week"

**Command**:
```bash
mac-whisper-tool search -t "standup" -t "team" -s 2025-12-26 --estimate-start -n 10
```

**What happens**:
- Searches for meetings with both "standup" AND "team" in title
- Filters to meetings from 2025-12-26 onwards
- Returns up to 10 results

### 3. Search by Date Range

**User request**: "Show all meetings from December"

**Command**:
```bash
mac-whisper-tool search -s 2025-12-01 -e 2025-12-31 --estimate-start -n 10
```

**What happens**:
- Lists meetings from December 2025
- No keyword filtering
- Returns up to 10 results

### 4. Find Meetings Mentioning a Client

**User request**: "Which meetings discussed the ABC client?"

**Command**:
```bash
mac-whisper-tool search -k "ABC" -k "client" --estimate-start -n 10
```

**What happens**:
- Searches for meetings containing both "ABC" AND "client"
- Returns up to 10 results

### 5. Get All Zoom Meetings

**User request**: "Find all Zoom meetings"

**Command**:
```bash
mac-whisper-tool search -t "zoom" --estimate-start -n -1
```

**What happens**:
- Searches for meetings with "zoom" in title
- Returns ALL results (no limit)

### 6. JSON Output for Processing

**User request**: "Get search results as JSON"

**Command**:
```bash
mac-whisper-tool search -k "project" --estimate-start -f json
```

**Output**:
```json
[
  {
    "sessionId": "ABC123XYZ",
    "dateCreated": "2025-12-15T15:45:30.000",
    "dateStarted": "2025-12-15T14:00:00.000",
    "duration": 2730.0,
    "title": "Project Planning Meeting",
    "preview": "Let's discuss the project timeline..."
  }
]
```

### 7. Combine Multiple Filters

**User request**: "Find budget meetings from Q4 2025"

**Command**:
```bash
mac-whisper-tool search -k "budget" -t "meeting" -s 2025-10-01 -e 2025-12-31 --estimate-start -n 10
```

**What happens**:
- Searches for "budget" in transcription
- Requires "meeting" in title
- Filters to Q4 2025 (Oct-Dec)
- Returns up to 10 results

## Workflow Examples

### Typical Search-to-Read Workflow

**Step 1**: Search for meetings
```bash
mac-whisper-tool search -k "design review" --estimate-start -n 10
```

**Output**:
```
Session ID         Start Time            Duration  Title           Preview
xY9zA1bC2d        2025-12-15 14:00      45m 30s   Design Review   We discussed the new...
```

**Step 2**: Use session ID with macwhisper-read skill
```bash
mac-whisper-tool export -x --estimate-start "xY9zA1bC2d"
```

**Result**: Full transcription with timestamps and metadata

## Troubleshooting

### No Results Found

**Problem**: Search returns no results

**Possible reasons**:
- Keywords are too specific (try broader terms)
- Date range doesn't match any meetings
- Keywords are case-sensitive in some contexts

**Solution**:
- Try fewer keywords
- Broaden date range
- Try title search instead of content search

### "At Least One Search Criterion Required" Error

**Problem**: Command fails with this error

**Cause**: No search parameters provided (no -k, -t, -s, or -e flags)

**Solution**: Add at least one search criterion:
```bash
mac-whisper-tool search -s 2025-01-01 --estimate-start -n 10
```

### Database Not Found Error

**Problem**: "failed to open database" error

**Cause**: Database file not at default location

**Solution**: Ensure MacWhisper is installed, or specify custom database path:
```bash
mac-whisper-tool search -k "meeting" --db /path/to/main.sqlite --estimate-start -n 10
```

## Tips

1. **Start broad, then narrow**: Begin with fewer keywords, then add more to refine results

2. **Use title search for known meeting types**: Title search is faster and more accurate for regular meetings (e.g., "standup", "sync")

3. **Combine with date ranges**: Narrow down results by adding date filters to keyword searches

4. **Use JSON for automation**: The JSON output format is perfect for scripts and automation

5. **Session IDs are persistent**: Save session IDs for frequently accessed meetings

## Advanced Examples

### Find Meetings Within Last 7 Days

```bash
mac-whisper-tool search -s $(date -v-7d +%Y-%m-%d) --estimate-start -n 10
```

### Search for Multiple Topics (OR Condition)

For OR searches, run multiple searches:

```bash
# Search for "budget" OR "finance"
mac-whisper-tool search -k "budget" --estimate-start -f json -n -1 > budget.json
mac-whisper-tool search -k "finance" --estimate-start -f json -n -1 > finance.json
```

### Get Full List of All Meetings

```bash
mac-whisper-tool search -s 2000-01-01 --estimate-start -n -1
```

This returns all meetings in the database.
