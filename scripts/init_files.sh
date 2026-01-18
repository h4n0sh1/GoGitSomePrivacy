#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Create all the source files
echo "Creating source files..."

# Already created: internal/models/commit.go

cat > internal/models/result.go << 'MODELRESULT'
package models

// PIIMatch represents a detected instance of PII in a commit.
type PIIMatch struct {
	Commit      Commit       `json:"commit"`
	PIIType     PIIType      `json:"pii_type"`
	Locations   []Location   `json:"locations"`
	Confidence  float64      `json:"confidence"`
	Context     string       `json:"context"`
}

// PIIType represents the type of personally identifiable information.
type PIIType string

const (
	PIITypeFullName  PIIType = "full_name"
	PIITypeFirstName PIIType = "first_name"
	PIITypeLastName  PIIType = "last_name"
	PIITypeEmail     PIIType = "email"
	PIITypePhone     PIIType = "phone"
)

// Location represents where PII was found in the commit.
type Location struct {
	Field    string `json:"field"`     // e.g., "message", "author_name", "diff"
	Line     int    `json:"line"`      // Line number if applicable
	Column   int    `json:"column"`    // Column number if applicable
	Matched  string `json:"matched"`   // The actual text that matched
}

// ScanResult represents the complete scan results for a user.
type ScanResult struct {
	Username      string      `json:"username"`
	SearchedRepos int         `json:"searched_repos"`
	TotalCommits  int         `json:"total_commits"`
	Matches       []PIIMatch  `json:"matches"`
	ScanDuration  string      `json:"scan_duration"`
	Errors        []ScanError `json:"errors,omitempty"`
}

// ScanError represents errors encountered during scanning.
type ScanError struct {
	Repository string `json:"repository,omitempty"`
	Message    string `json:"message"`
	Severity   string `json:"severity"` // "warning", "error", "fatal"
}
MODELRESULT

cat > internal/models/user.go << 'MODELUSER'
package models

// UserProfile represents a GitHub user's profile information.
type UserProfile struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	AvatarURL string `json:"avatar_url"`
}

// PIISearchCriteria defines what PII to search for.
type PIISearchCriteria struct {
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	FullName      string   `json:"full_name"`
	Emails        []string `json:"emails,omitempty"`
	CaseSensitive bool     `json:"case_sensitive"`
}
MODELUSER

echo "Models created successfully!"
echo "Run: chmod +x scripts/init_files.sh && ./scripts/init_files.sh"
