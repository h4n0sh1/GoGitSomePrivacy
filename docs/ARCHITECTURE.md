# Architecture Documentation

## Overview

GoGitSomePrivacy is designed as a high-performance, concurrent application for scanning GitHub commits to detect personally identifiable information (PII). The architecture follows Go best practices and emphasizes modularity, testability, and performance.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│                   (cmd/gogitsomeprivacy)                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Commands & Flags
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   Configuration Layer                        │
│                    (internal/config)                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Config
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Scanner Service                           │
│                   (internal/scanner)                         │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ • Orchestrates scanning workflow                      │  │
│  │ • Manages worker pool                                 │  │
│  │ • Aggregates results                                  │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────┬──────────────────────────────────────┬────────────┘
          │                                      │
          │                                      │
┌─────────▼────────────┐              ┌─────────▼────────────┐
│   GitHub Client      │              │    Worker Pool       │
│ (internal/github)    │              │ (internal/worker)    │
│ ┌──────────────────┐ │              │ ┌──────────────────┐ │
│ │ • API calls      │ │              │ │ • Job queue      │ │
│ │ • Rate limiting  │ │              │ │ • Concurrency    │ │
│ │ • Error handling │ │              │ │ • Lifecycle mgmt │ │
│ └──────────────────┘ │              │ └──────────────────┘ │
└──────────────────────┘              └──────────────────────┘
          │                                      │
          │                                      │
          └──────────────┬───────────────────────┘
                         │
                         │ Process commits
                         │
                ┌────────▼────────┐
                │  PII Detector   │
                │   (pkg/pii)     │
                │ ┌──────────────┐│
                │ │ • Pattern    ││
                │ │   matching   ││
                │ │ • Confidence ││
                │ │   scoring    ││
                │ └──────────────┘│
                └─────────────────┘
```

## Component Details

### 1. CLI Layer (`cmd/gogitsomeprivacy`)

**Responsibility**: User interface and command execution

- Uses Cobra for command-line parsing
- Validates user input
- Formats and displays output
- Handles errors gracefully

**Key Features**:
- Multiple output formats (JSON, text)
- Configuration file support
- Environment variable overrides
- Verbose logging option

### 2. Configuration (`internal/config`)

**Responsibility**: Application configuration management

- Uses Viper for flexible configuration
- Supports multiple configuration sources:
  - YAML configuration files
  - Environment variables
  - Command-line flags
- Validates configuration values

**Configuration Hierarchy**:
1. Default values
2. Configuration file
3. Environment variables
4. Command-line flags (highest priority)

### 3. Scanner Service (`internal/scanner`)

**Responsibility**: Core scanning orchestration

**Workflow**:
1. Fetch user profile
2. List all public repositories
3. For each repository (concurrent):
   - Fetch commits by user
   - Scan each commit for PII
   - Collect matches
4. Aggregate results

**Concurrency Model**:
- Uses worker pool for repository scanning
- Channels for result collection
- Context for cancellation
- Structured error handling

### 4. GitHub Client (`internal/github`)

**Responsibility**: GitHub API interaction

**Features**:
- Wraps `go-github` library
- Token-based authentication
- Rate limiting (using `golang.org/x/time/rate`)
- Automatic pagination
- Error handling and retries

**Rate Limiting Strategy**:
- Configurable requests per second
- Token bucket algorithm
- Respects GitHub's rate limit headers

### 5. Worker Pool (`internal/worker`)

**Responsibility**: Concurrent job processing

**Design**:
- Fixed number of worker goroutines
- Job queue using channels
- Graceful shutdown
- Context-aware cancellation

**Benefits**:
- Prevents resource exhaustion
- Predictable memory usage
- Efficient CPU utilization
- Easy to test and monitor

### 6. PII Detector (`pkg/pii`)

**Responsibility**: PII detection in text

**Algorithm**:
- String matching with word boundaries
- Case-sensitive/insensitive search
- Line and column tracking
- Confidence scoring

**Word Boundary Detection**:
- Ensures matches are complete words
- Prevents false positives (e.g., "Johnson" vs "John")
- Uses Unicode-aware character classification

## Data Flow

### Scanning Process

```
User Input → Config → Scanner → GitHub Client → API
                ↓
         Worker Pool
                ↓
          PII Detector
                ↓
          Result Aggregation
                ↓
         Output Formatter
                ↓
             User
```

### Data Models

The application uses well-defined data models in `internal/models`:

- **UserProfile**: GitHub user information
- **Repository**: Repository metadata
- **Commit**: Commit information with author details
- **PIIMatch**: Detected PII with location and confidence
- **ScanResult**: Complete scan results with statistics
- **PIISearchCriteria**: Search parameters

## Concurrency Patterns

### Worker Pool Pattern

```go
// Create pool
pool := worker.NewPool(maxWorkers)
pool.Start(ctx, workerFunc)

// Submit jobs
for _, job := range jobs {
    pool.Submit(job)
}

// Wait for completion
pool.Stop()
pool.Wait()
```

### Result Collection

```go
// Buffered channels for results
matchesChan := make(chan PIIMatch, 100)
errorsChan := make(chan ScanError, 100)

// Collector goroutine
go func() {
    for match := range matchesChan {
        results = append(results, match)
    }
}()
```

## Error Handling

### Strategy

1. **Graceful Degradation**: Repository scan failures don't stop the entire process
2. **Error Collection**: All errors are collected and reported
3. **Severity Levels**: Warnings vs errors vs fatal
4. **Context Preservation**: Errors include repository and operation context

### Example

```go
if err := scanRepo(repo); err != nil {
    errorsChan <- ScanError{
        Repository: repo.FullName,
        Message:    err.Error(),
        Severity:   "warning",
    }
    return nil // Continue with other repos
}
```

## Testing Strategy

### Unit Tests

- Each package has comprehensive tests
- Test files follow `*_test.go` convention
- Table-driven tests for multiple scenarios
- Mocking external dependencies

### Benchmark Tests

- Performance-critical code has benchmarks
- Located in `*_test.go` files
- Run with `go test -bench=.`

### Test Organization

```
package/
├── implementation.go
└── implementation_test.go
```

## Performance Considerations

### Optimizations

1. **Concurrent Repository Scanning**: Multiple repositories processed in parallel
2. **Efficient String Matching**: Single-pass detection with word boundaries
3. **Rate Limiting**: Prevents API throttling
4. **Memory Management**: Bounded channels and worker pools

### Scalability

- **Horizontal**: Worker count can be increased
- **Vertical**: Efficient memory usage per worker
- **API Limits**: Respects and adapts to rate limits

### Benchmarks

Typical performance metrics:
- PII detection: ~1-2 µs per text scan
- Worker pool overhead: ~100 ns per job
- API call throughput: Up to 5000/hour (with token)

## Security

### Token Management

- Never logged or printed
- Stored in config files with restricted permissions
- Environment variable support for CI/CD

### API Access

- Read-only operations
- Public data only
- No data persistence

## Future Enhancements

### Planned Features

1. **Extended PII Types**: Email, phone, SSN patterns
2. **Diff Analysis**: Scan actual code changes
3. **Report Formats**: HTML, PDF, CSV
4. **Database Storage**: Persistent result storage
5. **Web UI**: Interactive result browsing
6. **Batch Mode**: Scan multiple users

### Architecture Extensions

- Plugin system for custom PII detectors
- Webhook support for CI/CD integration
- REST API for programmatic access
- Distributed scanning for large-scale operations

## Dependencies

### Core Dependencies

- `github.com/google/go-github/v58`: GitHub API client
- `github.com/spf13/cobra`: CLI framework
- `github.com/spf13/viper`: Configuration management
- `golang.org/x/oauth2`: OAuth2 authentication
- `golang.org/x/time/rate`: Rate limiting
- `golang.org/x/sync`: Concurrency utilities

### Development Dependencies

- `golangci-lint`: Code linting
- Standard Go testing tools

## Build and Deployment

### Build Targets

- Multiple OS/architecture combinations
- Optimized binary size with `-ldflags "-s -w"`
- Version information embedded at build time

### Deployment Options

1. **Binary Distribution**: Static binaries for each platform
2. **Go Install**: `go install` from source
3. **Container**: Docker image (future)
4. **Package Managers**: Homebrew, apt, etc. (future)

## Monitoring and Observability

### Logging

- Structured logging with severity levels
- Progress reporting in verbose mode
- Error context preservation

### Metrics

- Scan duration
- Repositories scanned
- Commits processed
- Matches found
- Error count by type

## Best Practices Applied

1. **Package Organization**: Clear separation between `internal` and `pkg`
2. **Dependency Management**: Go modules with version pinning
3. **Error Handling**: Explicit error returns, no panics in library code
4. **Context Usage**: Cancellation and timeout support
5. **Testing**: Comprehensive test coverage
6. **Documentation**: Inline comments and external docs
7. **Code Style**: Follows `gofmt` and linter rules
8. **API Design**: Intuitive interfaces, minimal coupling
