package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"devpulse/internal/domain"
)

type Client interface {
	FetchRecentCommits(ctx context.Context, repoName, branch, author, token string, since time.Time, loc *time.Location) ([]*domain.WorkLog, error)
	CheckRepoAccess(ctx context.Context, repoName, token string) (bool, error)
}

type HTTPClient struct {
	httpClient *http.Client
}

func NewClient() *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type gitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

func (c *HTTPClient) FetchRecentCommits(ctx context.Context, repoName, branch, author, token string, since time.Time, loc *time.Location) ([]*domain.WorkLog, error) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s", parts[0], parts[1], branch)
	if !since.IsZero() {
		url += fmt.Sprintf("&since=%s", since.UTC().Format(time.RFC3339))
	}
	if author != "" {
		url += fmt.Sprintf("&author=%s", author)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "DevPulse-Microservice")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository %s not found or inaccessible", repoName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API error (status %d)", resp.StatusCode)
	}

	var commits []gitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("failed to parse github commits: %w", err)
	}

	var logs []*domain.WorkLog
	for _, commit := range commits {
		fullMsg := strings.TrimSpace(commit.Commit.Message)
		lines := strings.Split(fullMsg, "\n")
		title := fullMsg
		if len(lines) > 0 {
			title = strings.TrimSpace(lines[0])
		}

		commitDate := commit.Commit.Author.Date.In(loc)
		dateOnly := time.Date(commitDate.Year(), commitDate.Month(), commitDate.Day(), 0, 0, 0, 0, loc)

		logs = append(logs, &domain.WorkLog{
			Category:   domain.CategoryCode,
			Source:     "GIT",
			RepoName:   repoName,
			Title:      title,
			Details:    fullMsg,
			CommitHash: commit.SHA,
			LogDate:    dateOnly,
			CreatedAt:  commitDate,
		})
	}

	return logs, nil
}

func (c *HTTPClient) CheckRepoAccess(ctx context.Context, repoName, token string) (bool, error) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", parts[0], parts[1])
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "DevPulse-Microservice")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("github API returned status %d", resp.StatusCode)
}
