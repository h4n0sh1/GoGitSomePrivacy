// Package pii provides PII detection utilities.
package pii

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/h4n0sh1/GoGitSomePrivacy/internal/models"
)

// Detector detects personally identifiable information in text.
type Detector struct {
	criteria      models.PIISearchCriteria
	patterns      map[models.PIIType]*regexp.Regexp
	caseSensitive bool
	contextSize   int
}

// NewDetector creates a new PII detector.
func NewDetector(criteria models.PIISearchCriteria, contextSize int) *Detector {
	d := &Detector{
		criteria:      criteria,
		patterns:      make(map[models.PIIType]*regexp.Regexp),
		caseSensitive: criteria.CaseSensitive,
		contextSize:   contextSize,
	}
	d.compilePatterns()
	return d
}

// compilePatterns compiles regex patterns for the search criteria.
func (d *Detector) compilePatterns() {
	flags := ""
	if !d.caseSensitive {
		flags = "(?i)"
	}

	// Full name pattern with word boundaries
	if d.criteria.FullName != "" {
		pattern := flags + `\b` + regexp.QuoteMeta(d.criteria.FullName) + `\b`
		if re, err := regexp.Compile(pattern); err == nil {
			d.patterns[models.PIITypeFullName] = re
		}
	}

	// First name pattern with word boundaries
	if d.criteria.FirstName != "" {
		pattern := flags + `\b` + regexp.QuoteMeta(d.criteria.FirstName) + `\b`
		if re, err := regexp.Compile(pattern); err == nil {
			d.patterns[models.PIITypeFirstName] = re
		}
	}

	// Last name pattern with word boundaries
	if d.criteria.LastName != "" {
		pattern := flags + `\b` + regexp.QuoteMeta(d.criteria.LastName) + `\b`
		if re, err := regexp.Compile(pattern); err == nil {
			d.patterns[models.PIITypeLastName] = re
		}
	}
}

// Match represents a single match found in text.
type Match struct {
	Type    models.PIIType
	Text    string
	Start   int
	End     int
	Context string
	Field   string
	Line    int
	Column  int
}

// DetectInCommit detects PII in a commit.
func (d *Detector) DetectInCommit(commit *models.Commit) []Match {
	var matches []Match

	// Check commit message
	msgMatches := d.detectInText(commit.Message, "message")
	matches = append(matches, msgMatches...)

	// Check author name
	if commit.Author.Name != "" {
		authorMatches := d.detectInText(commit.Author.Name, "author_name")
		matches = append(matches, authorMatches...)
	}

	// Check committer name
	if commit.Committer.Name != "" && commit.Committer.Name != commit.Author.Name {
		committerMatches := d.detectInText(commit.Committer.Name, "committer_name")
		matches = append(matches, committerMatches...)
	}

	return matches
}

// detectInText detects PII in a text string.
func (d *Detector) detectInText(text, field string) []Match {
	var matches []Match

	for piiType, pattern := range d.patterns {
		if pattern == nil {
			continue
		}

		allMatches := pattern.FindAllStringIndex(text, -1)
		for _, loc := range allMatches {
			start, end := loc[0], loc[1]
			matchedText := text[start:end]

			// Calculate line and column
			line, col := d.getLineCol(text, start)

			// Extract context
			context := d.extractContext(text, start, end)

			matches = append(matches, Match{
				Type:    piiType,
				Text:    matchedText,
				Start:   start,
				End:     end,
				Context: context,
				Field:   field,
				Line:    line,
				Column:  col,
			})
		}
	}

	return matches
}

// getLineCol calculates line and column numbers for a position.
func (d *Detector) getLineCol(text string, pos int) (int, int) {
	line := 1
	lastNewline := 0

	for i := 0; i < pos && i < len(text); i++ {
		if text[i] == '\n' {
			line++
			lastNewline = i + 1
		}
	}

	return line, pos - lastNewline + 1
}

// extractContext extracts surrounding context for a match.
func (d *Detector) extractContext(text string, start, end int) string {
	ctxStart := start - d.contextSize
	if ctxStart < 0 {
		ctxStart = 0
	}

	ctxEnd := end + d.contextSize
	if ctxEnd > len(text) {
		ctxEnd = len(text)
	}

	// Trim to word boundaries
	ctx := text[ctxStart:ctxEnd]

	// Clean up whitespace
	ctx = strings.Join(strings.Fields(ctx), " ")

	return ctx
}

// CalculateConfidence calculates a confidence score for matches.
func CalculateConfidence(matches []Match) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	// Base confidence
	confidence := 0.7

	// More matches = higher confidence
	if len(matches) > 1 {
		confidence += 0.05 * float64(min(len(matches)-1, 3))
	}

	// Full name match is higher confidence
	for _, m := range matches {
		if m.Type == models.PIITypeFullName {
			confidence += 0.1
			break
		}
	}

	// Matches in author field are higher confidence
	for _, m := range matches {
		if m.Field == "author_name" || m.Field == "committer_name" {
			confidence += 0.05
			break
		}
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// IsLikelyFalsePositive checks if a match is likely a false positive.
func IsLikelyFalsePositive(match Match, text string) bool {
	// Check if the match is part of a larger word (shouldn't happen with word boundaries, but double-check)
	if match.Start > 0 {
		prevChar := rune(text[match.Start-1])
		if unicode.IsLetter(prevChar) || unicode.IsDigit(prevChar) {
			return true
		}
	}

	if match.End < len(text) {
		nextChar := rune(text[match.End])
		if unicode.IsLetter(nextChar) || unicode.IsDigit(nextChar) {
			return true
		}
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
