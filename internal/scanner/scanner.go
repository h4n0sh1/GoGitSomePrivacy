// Package scanner provides the main scanning logic for PII detection.
package scanner

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/h4n0sh1/GoGitSomePrivacy/internal/github"
	"github.com/h4n0sh1/GoGitSomePrivacy/internal/models"
	"github.com/h4n0sh1/GoGitSomePrivacy/internal/worker"
	"github.com/h4n0sh1/GoGitSomePrivacy/pkg/pii"
)

// Config contains scanner configuration.
type Config struct {
	MaxWorkers     int
	ContextSize    int
	ProgressLogger *log.Logger
}

// Scanner scans GitHub commits for PII.
type Scanner struct {
	client   *github.Client
	criteria models.PIISearchCriteria
	config   Config
	detector *pii.Detector
}

// NewScanner creates a new scanner.
func NewScanner(client *github.Client, criteria models.PIISearchCriteria, config Config) *Scanner {
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 10
	}
	if config.ContextSize <= 0 {
		config.ContextSize = 50
	}

	return &Scanner{
		client:   client,
		criteria: criteria,
		config:   config,
		detector: pii.NewDetector(criteria, config.ContextSize),
	}
}

// repoCommits holds commits for a repository.
type repoCommits struct {
	Repo    *models.Repository
	Commits []*models.Commit
	Err     error
}

// ScanUser scans all commits by a user for PII.
func (s *Scanner) ScanUser(ctx context.Context, username string) (*models.ScanResult, error) {
	startTime := time.Now()

	result := &models.ScanResult{
		Username: username,
		Matches:  []models.PIIMatch{},
		Errors:   []models.ScanError{},
	}

	s.log("Starting scan for user: %s", username)

	// Get user profile
	profile, err := s.client.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	s.log("Found user: %s (%s)", profile.Login, profile.Name)

	// List all repositories
	s.log("Fetching repositories...")
	repos, err := s.client.ListUserRepos(ctx, username)
	if err != nil {
		return nil, err
	}
	result.SearchedRepos = len(repos)
	s.log("Found %d public repositories", len(repos))

	// Create worker pool
	pool := worker.NewPool(s.config.MaxWorkers, func(ctx context.Context, repo *models.Repository) (*repoCommits, error) {
		commits, err := s.client.ListUserCommits(ctx, repo.Owner, repo.Name, username)
		return &repoCommits{Repo: repo, Commits: commits, Err: err}, nil
	})

	// Start workers
	pool.Start(ctx)

	// Submit repos to pool
	go func() {
		for _, repo := range repos {
			pool.Submit(repo)
		}
		pool.Close()
	}()

	// Collect results and scan for PII
	var mu sync.Mutex
	var totalCommits int

	for task := range pool.Results() {
		if task.Err != nil {
			mu.Lock()
			result.Errors = append(result.Errors, models.ScanError{
				Repository: task.Result.Repo.FullName,
				Message:    task.Err.Error(),
				Severity:   "warning",
			})
			mu.Unlock()
			continue
		}

		rc := task.Result
		if rc.Err != nil {
			mu.Lock()
			result.Errors = append(result.Errors, models.ScanError{
				Repository: rc.Repo.FullName,
				Message:    rc.Err.Error(),
				Severity:   "warning",
			})
			mu.Unlock()
			continue
		}

		s.log("Scanning %d commits in %s", len(rc.Commits), rc.Repo.FullName)

		for _, commit := range rc.Commits {
			totalCommits++

			// Detect PII
			matches := s.detector.DetectInCommit(commit)
			if len(matches) > 0 {
				piiMatch := s.buildPIIMatch(commit, matches)
				mu.Lock()
				result.Matches = append(result.Matches, piiMatch)
				mu.Unlock()
			}
		}
	}

	result.TotalCommits = totalCommits
	result.ScanDuration = time.Since(startTime).String()

	s.log("Scan complete: %d commits, %d matches, duration: %s",
		result.TotalCommits, len(result.Matches), result.ScanDuration)

	return result, nil
}

// buildPIIMatch builds a PIIMatch from detected matches.
func (s *Scanner) buildPIIMatch(commit *models.Commit, matches []pii.Match) models.PIIMatch {
	locations := make([]models.Location, len(matches))
	for i, m := range matches {
		locations[i] = models.Location{
			Field:   m.Field,
			Line:    m.Line,
			Column:  m.Column,
			Matched: m.Text,
		}
	}

	// Use the first match's type as the primary type
	piiType := models.PIITypeFullName
	if len(matches) > 0 {
		piiType = matches[0].Type
	}

	// Get context from first match
	context := ""
	if len(matches) > 0 {
		context = matches[0].Context
	}

	return models.PIIMatch{
		Commit:     *commit,
		PIIType:    piiType,
		Locations:  locations,
		Confidence: pii.CalculateConfidence(matches),
		Context:    context,
	}
}

// log logs a message if verbose logging is enabled.
func (s *Scanner) log(format string, args ...interface{}) {
	if s.config.ProgressLogger != nil {
		s.config.ProgressLogger.Printf(format, args...)
	}
}
