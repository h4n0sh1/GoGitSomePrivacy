package models

import "time"

// Commit represents a Git commit with relevant information for PII scanning.
type Commit struct {
	SHA        string    `json:"sha"`
	Repository string    `json:"repository"`
	Message    string    `json:"message"`
	Author     Author    `json:"author"`
	Committer  Author    `json:"committer"`
	Date       time.Time `json:"date"`
	URL        string    `json:"url"`
}

// Author represents commit author information.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Login string `json:"login"`
}

// Repository represents a GitHub repository.
type Repository struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Private     bool   `json:"private"`
}
