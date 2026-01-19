// Package config provides configuration management for GoGitSomePrivacy.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	GitHub GitHubConfig `yaml:"github"`
	Scan   ScanConfig   `yaml:"scan"`
}

// GitHubConfig contains GitHub API settings.
type GitHubConfig struct {
	Token              string  `yaml:"token"`
	RateLimitPerSecond float64 `yaml:"rate_limit_per_second"`
	TimeoutSeconds     int     `yaml:"timeout_seconds"`
}

// ScanConfig contains scanning settings.
type ScanConfig struct {
	MaxWorkers       int  `yaml:"max_workers"`
	ContextSize      int  `yaml:"context_size"`
	CaseSensitive    bool `yaml:"case_sensitive"`
	IncludeAuthor    bool `yaml:"include_author"`
	IncludeCommitter bool `yaml:"include_committer"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		GitHub: GitHubConfig{
			Token:              "",
			RateLimitPerSecond: 1.3,
			TimeoutSeconds:     30,
		},
		Scan: ScanConfig{
			MaxWorkers:       10,
			ContextSize:      50,
			CaseSensitive:    false,
			IncludeAuthor:    true,
			IncludeCommitter: true,
		},
	}
}

// Load loads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from config file
	if configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Try default locations
		defaultPaths := []string{
			filepath.Join(os.Getenv("HOME"), ".config", "gogitsomeprivacy", "config.yaml"),
			filepath.Join(os.Getenv("HOME"), ".config", "gogitsomeprivacy", "config.yml"),
			"config.yaml",
			"config.yml",
		}
		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				if err := loadFromFile(cfg, path); err == nil {
					break
				}
			}
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	return cfg, nil
}

func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func loadFromEnv(cfg *Config) {
	// GitHub token from environment
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cfg.GitHub.Token = token
	}
	if token := os.Getenv("GGSP_GITHUB_TOKEN"); token != "" {
		cfg.GitHub.Token = token
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Scan.MaxWorkers < 1 {
		return fmt.Errorf("max_workers must be at least 1")
	}
	if c.GitHub.RateLimitPerSecond <= 0 {
		return fmt.Errorf("rate_limit_per_second must be positive")
	}
	if c.GitHub.TimeoutSeconds < 1 {
		return fmt.Errorf("timeout_seconds must be at least 1")
	}
	return nil
}
