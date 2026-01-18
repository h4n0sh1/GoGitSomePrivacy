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
