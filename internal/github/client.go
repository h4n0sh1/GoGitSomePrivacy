// Package github provides a GitHub API client for GoGitSomePrivacy.
package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v58/github"
	"github.com/h4n0sh1/GoGitSomePrivacy/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

// ClientConfig contains configuration for the GitHub client.
type ClientConfig struct {
	Token              string
	RateLimitPerSecond float64
	Timeout            time.Duration
}

// Client wraps the GitHub API client with rate limiting.
type Client struct {
	client      *github.Client
	rateLimiter *rate.Limiter
	timeout     time.Duration
}

// NewClient creates a new GitHub API client.
func NewClient(cfg ClientConfig) *Client {
	var httpClient *http.Client

	if cfg.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.Token},
		)
		httpClient = oauth2.NewClient(context.Background(), ts)
	} else {
		httpClient = http.DefaultClient
	}

	if cfg.Timeout > 0 {
		httpClient.Timeout = cfg.Timeout
	} else {
		httpClient.Timeout = 30 * time.Second
	}

	// Create rate limiter
	rps := cfg.RateLimitPerSecond
	if rps <= 0 {
		rps = 1.0 // Default: 1 request per second
	}
	limiter := rate.NewLimiter(rate.Limit(rps), 1)

	return &Client{
		client:      github.NewClient(httpClient),
		rateLimiter: limiter,
		timeout:     cfg.Timeout,
	}
}

// wait waits for rate limiter before making a request.
func (c *Client) wait(ctx context.Context) error {
	return c.rateLimiter.Wait(ctx)
}

// GetUser retrieves a GitHub user's profile.
func (c *Client) GetUser(ctx context.Context, username string) (*models.UserProfile, error) {
	if err := c.wait(ctx); err != nil {
		return nil, err
	}

	user, _, err := c.client.Users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}

	return &models.UserProfile{
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		Bio:       user.GetBio(),
		Company:   user.GetCompany(),
		Location:  user.GetLocation(),
		AvatarURL: user.GetAvatarURL(),
	}, nil
}

// ListUserRepos lists all public repositories for a user (owned, member, collaborator).
func (c *Client) ListUserRepos(ctx context.Context, username string) ([]*models.Repository, error) {
	var allRepos []*models.Repository
	opts := &github.RepositoryListOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		if err := c.wait(ctx); err != nil {
			return nil, err
		}

		repos, resp, err := c.client.Repositories.List(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for %s: %w", username, err)
		}

		for _, repo := range repos {
			if repo.GetPrivate() {
				continue
			}
			allRepos = append(allRepos, &models.Repository{
				FullName:    repo.GetFullName(),
				Name:        repo.GetName(),
				Owner:       repo.GetOwner().GetLogin(),
				Description: repo.GetDescription(),
				URL:         repo.GetHTMLURL(),
				Private:     repo.GetPrivate(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// ListUserCommits lists all commits by a user in a repository.
func (c *Client) ListUserCommits(ctx context.Context, owner, repo, username string) ([]*models.Commit, error) {
	var allCommits []*models.Commit
	opts := &github.CommitsListOptions{
		Author:      username,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		if err := c.wait(ctx); err != nil {
			return nil, err
		}

		commits, resp, err := c.client.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			// Skip repos we can't access
			if _, ok := err.(*github.ErrorResponse); ok {
				break
			}
			return nil, fmt.Errorf("failed to list commits in %s/%s: %w", owner, repo, err)
		}

		for _, commit := range commits {
			c := convertCommit(commit, owner, repo)
			if c != nil {
				allCommits = append(allCommits, c)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}

// SearchUserCommits searches for commits by a user across GitHub.
func (c *Client) SearchUserCommits(ctx context.Context, username string) ([]*models.Commit, error) {
	var allCommits []*models.Commit
	query := fmt.Sprintf("author:%s", username)
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		if err := c.wait(ctx); err != nil {
			return nil, err
		}

		result, resp, err := c.client.Search.Commits(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to search commits for %s: %w", username, err)
		}

		for _, commit := range result.Commits {
			repoOwner := ""
			repoName := ""
			if commit.Repository != nil {
				repoOwner = commit.Repository.GetOwner().GetLogin()
				repoName = commit.Repository.GetName()
			}
			c := convertCommitResult(commit, repoOwner, repoName)
			if c != nil {
				allCommits = append(allCommits, c)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}

func convertCommit(rc *github.RepositoryCommit, owner, repo string) *models.Commit {
	if rc == nil || rc.Commit == nil {
		return nil
	}

	commit := &models.Commit{
		SHA:        rc.GetSHA(),
		Repository: fmt.Sprintf("%s/%s", owner, repo),
		Message:    rc.Commit.GetMessage(),
		URL:        rc.GetHTMLURL(),
	}

	if rc.Commit.Author != nil {
		commit.Author = models.Author{
			Name:  rc.Commit.Author.GetName(),
			Email: rc.Commit.Author.GetEmail(),
		}
		if rc.Commit.Author.Date != nil {
			commit.Date = rc.Commit.Author.Date.Time
		}
	}
	if rc.Author != nil {
		commit.Author.Login = rc.Author.GetLogin()
	}

	if rc.Commit.Committer != nil {
		commit.Committer = models.Author{
			Name:  rc.Commit.Committer.GetName(),
			Email: rc.Commit.Committer.GetEmail(),
		}
	}
	if rc.Committer != nil {
		commit.Committer.Login = rc.Committer.GetLogin()
	}

	return commit
}

func convertCommitResult(cr *github.CommitResult, owner, repo string) *models.Commit {
	if cr == nil || cr.Commit == nil {
		return nil
	}

	commit := &models.Commit{
		SHA:        cr.GetSHA(),
		Repository: fmt.Sprintf("%s/%s", owner, repo),
		Message:    cr.Commit.GetMessage(),
		URL:        cr.GetHTMLURL(),
	}

	if cr.Commit.Author != nil {
		commit.Author = models.Author{
			Name:  cr.Commit.Author.GetName(),
			Email: cr.Commit.Author.GetEmail(),
		}
		if cr.Commit.Author.Date != nil {
			commit.Date = cr.Commit.Author.Date.Time
		}
	}
	if cr.Author != nil {
		commit.Author.Login = cr.Author.GetLogin()
	}

	if cr.Commit.Committer != nil {
		commit.Committer = models.Author{
			Name:  cr.Commit.Committer.GetName(),
			Email: cr.Commit.Committer.GetEmail(),
		}
	}
	if cr.Committer != nil {
		commit.Committer.Login = cr.Committer.GetLogin()
	}

	return commit
}
