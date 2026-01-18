# Usage Guide

## Quick Start

### 1. Get a GitHub Token (Optional but Recommended)

While you can use the tool without authentication, GitHub's rate limits for unauthenticated requests are very low (60 requests/hour). With a token, you get 5000 requests/hour.

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Give it a name (e.g., "GoGitSomePrivacy")
4. Select scopes: `public_repo` (read-only access to public repositories)
5. Generate and copy the token

### 2. Set Up Your Token

```bash
# Option 1: Environment variable
export GITHUB_TOKEN="ghp_your_token_here"

# Option 2: Configuration file
mkdir -p ~/.config/gogitsomeprivacy
cat > ~/.config/gogitsomeprivacy/config.yaml << EOF
github:
  token: "ghp_your_token_here"
EOF
```

### 3. Run Your First Scan

```bash
# Simple - automatically searches for "The", "Octocat", and "The Octocat"
gogitsomeprivacy scan octocat --full-name "The Octocat"

# Or search for exact phrase only
gogitsomeprivacy scan octocat --full-name "The Octocat" --exact
```

## Common Use Cases

### Scanning for Full Name

```bash
# Smart search (default) - searches for:
# 1. "John Doe" (full name)
# 2. "John" (first name)
# 3. "Doe" (last name)
gogitsomeprivacy scan username --full-name "John Doe"

# Exact match only - only searches for "John Doe" as exact phrase
gogitsomeprivacy scan username --full-name "John Doe" --exact

# Case-sensitive search
gogitsomeprivacy scan username --full-name "John Doe" --case-sensitive

# Save to file
gogitsomeprivacy scan username --full-name "John Doe" -f results.json
```

### Scanning for First and Last Names Separately

```bash
# Manual control - override auto-detection
gogitsomeprivacy scan username --first-name "John" --last-name "Doe"

# Search only first name
gogitsomeprivacy scan username --first-name "John"

# Search only last name  
gogitsomeprivacy scan username --last-name "Doe"

# Combine with full name for maximum coverage
gogitsomeprivacy scan username \
  --full-name "John Doe" \
  --first-name "Johnny" \
  --last-name "Doe-Smith"
```

### Performance Tuning

```bash
# Increase workers for faster scanning
gogitsomeprivacy scan username --full-name "John Doe" --workers 20

# Reduce workers if hitting rate limits
gogitsomeprivacy scan username --full-name "John Doe" --workers 5
```

### Output Formats

```bash
# JSON output (default)
gogitsomeprivacy scan username --full-name "John Doe" -o json

# Human-readable text
gogitsomeprivacy scan username --full-name "John Doe" -o text

# Save to file
gogitsomeprivacy scan username --full-name "John Doe" -o json -f results.json
gogitsomeprivacy scan username --full-name "John Doe" -o text -f report.txt
```

### Verbose Output

```bash
# See progress and debugging information
gogitsomeprivacy scan username --full-name "John Doe" --verbose
```

## Understanding Results

### JSON Output Structure

```json
{
  "username": "octocat",
  "searched_repos": 8,
  "total_commits": 0,
  "matches": [
    {
      "commit": {
        "sha": "abc123...",
        "repository": "owner/repo",
        "message": "Fix bug reported by John Doe",
        "author": {
          "name": "jdoe",
          "email": "jdoe@example.com",
          "login": "jdoe"
        },
        "date": "2024-01-15T10:30:00Z",
        "url": "https://github.com/owner/repo/commit/abc123"
      },
      "pii_type": "full_name",
      "locations": [
        {
          "field": "message",
          "line": 1,
          "column": 20,
          "matched": "John Doe"
        }
      ],
      "confidence": 0.75,
      "context": "Fix bug reported by John Doe for..."
    }
  ],
  "scan_duration": "2m34.5s",
  "errors": []
}
```

### Confidence Scores

- **0.7 - 0.75**: Single match, medium confidence
- **0.75 - 0.85**: Multiple matches in same commit
- **0.85 - 1.0**: Many matches across different fields

### Location Fields

- `message`: Found in commit message
- `author_name`: Found in commit author name
- `committer_name`: Found in committer name

### Text Output Example

```
Scan Results for: octocat
=============================

Repositories Scanned: 8
Total Commits: 45
PII Matches Found: 3
Scan Duration: 2m34.5s

Matches:
--------

1. Repository: owner/repo
   Commit: abc12345
   Date: 2024-01-15T10:30:00Z
   URL: https://github.com/owner/repo/commit/abc12345
   Confidence: 0.75
   Locations: 1 match(es)
     - Field: message, Match: "John Doe"
   Context: Fix bug reported by John Doe for testing
```

## Advanced Configuration

### Configuration File

Create `~/.config/gogitsomeprivacy/config.yaml`:

```yaml
github:
  # Your GitHub token
  token: "ghp_your_token_here"
  
  # API requests per second (adjust based on your needs)
  rate_limit_per_second: 10.0
  
  # Timeout for API requests
  timeout_seconds: 30

scan:
  # Number of concurrent workers
  max_workers: 10
  
  # Characters of context around matches
  context_size: 50
  
  # Case-sensitive matching
  case_sensitive: false
  
  # Include author name in search
  include_author: true
  
  # Include committer name in search
  include_committer: true
```

### Environment Variables

All configuration values can be set via environment variables:

```bash
# GitHub settings
export GGSP_GITHUB_TOKEN="ghp_your_token_here"
export GGSP_GITHUB_RATE_LIMIT_PER_SECOND="15.0"
export GGSP_GITHUB_TIMEOUT_SECONDS="60"

# Scan settings
export GGSP_SCAN_MAX_WORKERS="20"
export GGSP_SCAN_CONTEXT_SIZE="100"
export GGSP_SCAN_CASE_SENSITIVE="true"

# Run scan
gogitsomeprivacy scan username --full-name "John Doe"
```

### Configuration Priority

The tool uses this priority order (highest to lowest):

1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

## Troubleshooting

### Rate Limit Errors

```
Error: rate limit exceeded
```

**Solution**:
- Use a GitHub token
- Reduce `--workers` count
- Increase `rate_limit_per_second` in config
- Wait for rate limit to reset (shown in error message)

### No Results Found

If you're not finding expected matches:

1. **Check spelling**: Ensure names are spelled correctly
2. **Try case-insensitive**: Don't use `--case-sensitive`
3. **Use separate names**: Try `--first-name` and `--last-name` separately
4. **Check visibility**: Only public repositories and commits are scanned
5. **Verify commits**: The user must be the author or committer

### Timeout Errors

```
Error: context deadline exceeded
```

**Solution**:
- Increase `timeout_seconds` in config
- Check your internet connection
- Reduce `--workers` count
- Try again later

### Memory Issues

For users with thousands of commits:

1. Reduce `--workers` count
2. Process in smaller batches
3. Increase available memory

## Performance Tips

### Optimal Worker Count

- **Small scans** (< 10 repos): 5-10 workers
- **Medium scans** (10-100 repos): 10-15 workers
- **Large scans** (> 100 repos): 15-20 workers

### Maximizing Speed

```bash
# Fast scan with higher concurrency
gogitsomeprivacy scan username \
  --full-name "John Doe" \
  --workers 20 \
  --verbose
```

### Staying Within Rate Limits

```bash
# Conservative scan
gogitsomeprivacy scan username \
  --full-name "John Doe" \
  --workers 5 \
  --config ~/.config/gogitsomeprivacy/config.yaml
```

With `config.yaml`:
```yaml
github:
  rate_limit_per_second: 5.0
```

## Scripting and Automation

### Batch Processing

```bash
#!/bin/bash
# Scan multiple users

users=("user1" "user2" "user3")
name="John Doe"

for user in "${users[@]}"; do
  echo "Scanning $user..."
  gogitsomeprivacy scan "$user" \
    --full-name "$name" \
    --output json \
    --file "results-${user}.json"
done
```

### CI/CD Integration

```yaml
# GitHub Actions example
name: PII Scan

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Install tool
        run: go install github.com/h4n0sh1/GoGitSomePrivacy/cmd/gogitsomeprivacy@latest
      
      - name: Run scan
        env:
          GITHUB_TOKEN: ${{ secrets.SCAN_TOKEN }}
        run: |
          gogitsomeprivacy scan ${{ github.repository_owner }} \
            --full-name "Your Name" \
            --output json \
            --file results.json
      
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: scan-results
          path: results.json
```

### Parsing JSON Results with jq

```bash
# Count total matches
jq '.matches | length' results.json

# List all affected repositories
jq -r '.matches[].commit.repository' results.json | sort -u

# Find matches with high confidence
jq '.matches[] | select(.confidence > 0.8)' results.json

# Extract commit URLs
jq -r '.matches[].commit.url' results.json

# Group by repository
jq 'group_by(.commit.repository) | 
    map({repo: .[0].commit.repository, count: length})' results.json
```

## Best Practices

1. **Always use a token**: Better rate limits and faster scans
2. **Start conservative**: Begin with fewer workers and increase if needed
3. **Use verbose mode**: Helps understand progress and troubleshoot issues
4. **Save results**: Always use `-f` to save results for later analysis
5. **Regular scans**: Set up periodic scans to catch new commits
6. **Review results**: High confidence doesn't mean 100% accuracy
7. **Respect rate limits**: Don't abuse the GitHub API

## Privacy Considerations

- Tool only accesses **public** data
- No data is stored by the tool
- No data is sent anywhere except GitHub's API
- Results are only saved where you specify
- Use responsibly and ethically
