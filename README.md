# GoGitSomePrivacy

A high-performance, concurrent tool for scanning GitHub commits to detect personally identifiable information (PII) across a user's public repository history.

## 🎯 Features

- 🔍 **Smart PII Detection**: Automatically searches for first name, last name, and full name combinations
- ⚡ **Concurrent Scanning**: Multi-threaded architecture with configurable worker pools for maximum speed
- 🎯 **Flexible Search**: Use `--full-name "John Doe"` to automatically search for "John", "Doe", and "John Doe"
- 📊 **Multiple Output Formats**: JSON and human-readable text output
- 🔒 **Rate Limiting**: Built-in GitHub API rate limiting to prevent quota exhaustion
- ⚙️ **Highly Configurable**: YAML config files, environment variables, and CLI flags

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/h4n0sh1/GoGitSomePrivacy.git
cd GoGitSomePrivacy

# Download dependencies
go mod download

# Build the binary
make build

# The binary will be at ./build/bin/gogitsomeprivacy
```

### Basic Usage

```bash
# Smart search - automatically finds "John", "Doe", and "John Doe"
gogitsomeprivacy scan username --full-name "John Doe"

# Exact match only - only finds "John Doe" as complete phrase
gogitsomeprivacy scan username --full-name "John Doe" --exact

# With GitHub token for better rate limits
export GITHUB_TOKEN="ghp_your_token_here"
gogitsomeprivacy scan username --full-name "John Doe" --verbose

# Maximum performance with 20 workers
gogitsomeprivacy scan username --full-name "John Doe" --workers 20 --verbose
```

## 📖 Usage Examples

### Simple Scan

```bash
# Scan a user's commits for PII
gogitsomeprivacy scan octocat --full-name "The Octocat"
```

**What this does:**
- ✅ Searches for "The Octocat" (full phrase)
- ✅ Searches for "The" (first name)
- ✅ Searches for "Octocat" (last name)
- ✅ Scans all public repositories
- ✅ Checks commit messages, author names, and committer names

### Advanced Options

```bash
# Manual name control
gogitsomeprivacy scan username \
  --first-name "John" \
  --last-name "Doe-Smith"

# Case-sensitive search
gogitsomeprivacy scan username \
  --full-name "John Doe" \
  --case-sensitive

# Save results to file
gogitsomeprivacy scan username \
  --full-name "John Doe" \
  --output json \
  --file results.json

# High-performance scan with verbose output
gogitsomeprivacy scan username \
  --full-name "John Doe" \
  --workers 20 \
  --token "ghp_your_token" \
  --verbose
```

## ⚙️ Configuration

### GitHub Token (Recommended)

Without a token: **60 requests/hour** ❌  
With a token: **5,000 requests/hour** ✅

```bash
# Set via environment variable
export GITHUB_TOKEN="ghp_your_token_here"

# Or via config file
mkdir -p ~/.config/gogitsomeprivacy
cat > ~/.config/gogitsomeprivacy/config.yaml << EOF
github:
  token: "ghp_your_token_here"
  rate_limit_per_second: 1.3
scan:
  max_workers: 20
EOF
```

### Performance Configuration

For maximum speed while respecting API limits:

```yaml
github:
  token: "ghp_your_token"
  rate_limit_per_second: 1.3  # Stay under GitHub's ~1.4/s limit
  timeout_seconds: 30

scan:
  max_workers: 20              # More workers = better parallelism
  context_size: 50
  case_sensitive: false
```

## 🎛️ Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--full-name` | Full name to search for (auto-splits into first/last) | - |
| `--first-name` | First name to search for | - |
| `--last-name` | Last name to search for | - |
| `--exact` | Only search exact full name (disable auto-split) | `false` |
| `--workers` | Number of concurrent workers | `10` |
| `--token` | GitHub API token | - |
| `--output, -o` | Output format (`json`, `text`) | `json` |
| `--file, -f` | Output file path | stdout |
| `--case-sensitive` | Perform case-sensitive search | `false` |
| `--verbose, -v` | Verbose output with progress | `false` |
| `--config, -c` | Config file path | - |

## 📊 Output Example

### JSON Output

```json
{
  "username": "octocat",
  "searched_repos": 8,
  "total_commits": 45,
  "matches": [
    {
      "commit": {
        "sha": "abc123...",
        "repository": "owner/repo",
        "message": "Fix bug reported by John Doe",
        "author": {
          "name": "jdoe",
          "email": "jdoe@example.com"
        },
        "date": "2024-01-15T10:30:00Z",
        "url": "https://github.com/owner/repo/commit/abc123"
      },
      "pii_type": "full_name",
      "locations": [
        {
          "field": "message",
          "matched": "John Doe"
        }
      ],
      "confidence": 0.75
    }
  ],
  "scan_duration": "2m34.5s"
}
```

### Text Output

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
   Confidence: 0.75
   Locations: 1 match(es)
     - Field: message, Match: "John Doe"
```

## 🏗️ Project Structure

```
GoGitSomePrivacy/
├── cmd/gogitsomeprivacy/      # CLI entry point
├── internal/                   # Private application code
│   ├── config/                 # Configuration management
│   ├── github/                 # GitHub API client
│   ├── models/                 # Data models
│   ├── scanner/                # Core scanning logic
│   └── worker/                 # Worker pool implementation
├── pkg/pii/                    # Public PII detection library
├── docs/                       # Documentation
│   ├── ARCHITECTURE.md         # System architecture
│   └── USAGE.md               # Detailed usage guide
├── Makefile                    # Build automation
└── README.md                   # This file
```

## 🛠️ Development

### Building

```bash
# Build for current platform
make build

# Run tests
make test

# Generate coverage report
make coverage

# Run linter
make lint

# Format code
make fmt
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/scanner/...
```

## 📚 Documentation

- **[Usage Guide](docs/USAGE.md)**: Complete usage instructions with examples
- **[Architecture](docs/ARCHITECTURE.md)**: Detailed system design and components
- **[Contributing](CONTRIBUTING.md)**: How to contribute to the project

## 🔒 Security & Privacy

- ✅ Only accesses **public** repositories and commits
- ✅ No data is stored or transmitted (except to GitHub's API)
- ✅ Results are only saved where you specify
- ✅ GitHub tokens are never logged
- ⚠️ Use responsibly and ethically

## 🚦 Performance Tips

### Optimal Worker Configuration

| Scan Size | Repositories | Recommended Workers |
|-----------|--------------|---------------------|
| Small | < 10 | 5-10 workers |
| Medium | 10-100 | 10-15 workers |
| Large | > 100 | 15-20 workers |

### Rate Limiting

- **Without token**: Limited to 60 req/hour → Not recommended
- **With token**: 5000 req/hour → Set `rate_limit_per_second: 1.3`
- **Respect limits**: Don't exceed 1.4 requests/second sustained

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📝 License

[Add your license here]

## 🙏 Credits

Built following Google's Go best practices and modern Go idioms.

## 📧 Support

For issues, questions, or contributions, please [open an issue](https://github.com/h4n0sh1/GoGitSomePrivacy/issues).
