package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/h4n0sh1/GoGitSomePrivacy/internal/config"
	"github.com/h4n0sh1/GoGitSomePrivacy/internal/github"
	"github.com/h4n0sh1/GoGitSomePrivacy/internal/models"
	"github.com/h4n0sh1/GoGitSomePrivacy/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "gogitsomeprivacy",
	Short: "Scan GitHub commits for personally identifiable information",
	Long: `GoGitSomePrivacy is a tool that scans all public commits made by a GitHub user
across all repositories they have participated in, searching for personally
identifiable information (PII) such as real names.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

var scanCmd = &cobra.Command{
	Use:   "scan [username]",
	Short: "Scan a GitHub user's commits for PII",
	Long: `Scan all public commits made by a GitHub user across all public repositories,
searching for personally identifiable information like their real name.`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

var (
	configFile    string
	firstName     string
	lastName      string
	fullName      string
	outputFormat  string
	outputFile    string
	githubToken   string
	maxWorkers    int
	caseSensitive bool
	exactMatch    bool
	verbose       bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	scanCmd.Flags().StringVar(&firstName, "first-name", "", "first name to search for")
	scanCmd.Flags().StringVar(&lastName, "last-name", "", "last name to search for")
	scanCmd.Flags().StringVar(&fullName, "full-name", "", "full name to search for (also searches first and last names unless --exact is used)")
	scanCmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "output format (json, text)")
	scanCmd.Flags().StringVarP(&outputFile, "file", "f", "", "output file (default: stdout)")
	scanCmd.Flags().StringVarP(&githubToken, "token", "t", "", "GitHub API token (overrides config)")
	scanCmd.Flags().IntVarP(&maxWorkers, "workers", "w", 0, "number of concurrent workers (overrides config)")
	scanCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "perform case-sensitive search")
	scanCmd.Flags().BoolVar(&exactMatch, "exact", false, "only search for exact full name (don't split into first/last)")

	rootCmd.AddCommand(scanCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with command-line flags
	if githubToken != "" {
		cfg.GitHub.Token = githubToken
	}
	if maxWorkers > 0 {
		cfg.Scan.MaxWorkers = maxWorkers
	}
	if caseSensitive {
		cfg.Scan.CaseSensitive = caseSensitive
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Auto-split full name into first and last names for better detection
	// unless --exact flag is used
	if fullName != "" && !exactMatch && firstName == "" && lastName == "" {
		parts := strings.Fields(fullName)
		if len(parts) >= 2 {
			firstName = parts[0]
			lastName = parts[len(parts)-1]
			if verbose {
				log.Printf("Auto-detecting: first name=%q, last name=%q (use --exact to disable)", firstName, lastName)
			}
		}
	}

	// Build search criteria
	criteria := models.PIISearchCriteria{
		FirstName:     firstName,
		LastName:      lastName,
		FullName:      fullName,
		CaseSensitive: cfg.Scan.CaseSensitive,
	}

	// Validate search criteria
	if criteria.FirstName == "" && criteria.LastName == "" && criteria.FullName == "" {
		return fmt.Errorf("at least one of --first-name, --last-name, or --full-name must be specified")
	}

	// Create GitHub client
	githubClient := github.NewClient(github.ClientConfig{
		Token:              cfg.GitHub.Token,
		RateLimitPerSecond: cfg.GitHub.RateLimitPerSecond,
		Timeout:            time.Duration(cfg.GitHub.TimeoutSeconds) * time.Second,
	})

	// Create scanner
	var progressLogger *log.Logger
	if verbose {
		progressLogger = log.New(os.Stderr, "[SCAN] ", log.LstdFlags)
	}

	scannerConfig := scanner.Config{
		MaxWorkers:     cfg.Scan.MaxWorkers,
		ContextSize:    cfg.Scan.ContextSize,
		ProgressLogger: progressLogger,
	}

	s := scanner.NewScanner(githubClient, criteria, scannerConfig)

	// Run scan
	ctx := context.Background()
	result, err := s.ScanUser(ctx, username)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Output results
	if err := outputResults(result, outputFormat, outputFile); err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	return nil
}

func outputResults(result *models.ScanResult, format, outputPath string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	case "text":
		output = []byte(formatTextOutput(result))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, output, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Results written to %s\n", outputPath)
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func formatTextOutput(result *models.ScanResult) string {
	var output string

	output += fmt.Sprintf("Scan Results for: %s\n", result.Username)
	output += fmt.Sprintf("====================%s\n\n", repeatChar('=', len(result.Username)))
	output += fmt.Sprintf("Repositories Scanned: %d\n", result.SearchedRepos)
	output += fmt.Sprintf("Total Commits: %d\n", result.TotalCommits)
	output += fmt.Sprintf("PII Matches Found: %d\n", len(result.Matches))
	output += fmt.Sprintf("Scan Duration: %s\n\n", result.ScanDuration)

	if len(result.Matches) > 0 {
		output += "Matches:\n"
		output += "--------\n\n"

		for i, match := range result.Matches {
			output += fmt.Sprintf("%d. Repository: %s\n", i+1, match.Commit.Repository)
			output += fmt.Sprintf("   Commit: %s\n", match.Commit.SHA[:8])
			output += fmt.Sprintf("   Date: %s\n", match.Commit.Date.Format(time.RFC3339))
			output += fmt.Sprintf("   URL: %s\n", match.Commit.URL)
			output += fmt.Sprintf("   Confidence: %.2f\n", match.Confidence)
			output += fmt.Sprintf("   Locations: %d match(es)\n", len(match.Locations))

			for _, loc := range match.Locations {
				output += fmt.Sprintf("     - Field: %s, Match: %q\n", loc.Field, loc.Matched)
			}

			if match.Context != "" {
				output += fmt.Sprintf("   Context: %s\n", match.Context)
			}
			output += "\n"
		}
	}

	if len(result.Errors) > 0 {
		output += "\nErrors:\n"
		output += "-------\n\n"

		for i, err := range result.Errors {
			output += fmt.Sprintf("%d. [%s] %s", i+1, err.Severity, err.Message)
			if err.Repository != "" {
				output += fmt.Sprintf(" (Repository: %s)", err.Repository)
			}
			output += "\n"
		}
	}

	return output
}

func repeatChar(char rune, count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = char
	}
	return string(result)
}
