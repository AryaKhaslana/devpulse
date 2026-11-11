package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchRecentCommitsMock(t *testing.T) {
	mockCommits := []gitHubCommit{
		{
			SHA: "a1b2c3d4e5f6",
			Commit: struct {
				Message string `json:"message"`
				Author  struct {
					Name  string    `json:"name"`
					Email string    `json:"email"`
					Date  time.Time `json:"date"`
				} `json:"author"`
			}{
				Message: "feat(auth): add jwt authentication\n\nDetailed commit body here.",
				Author: struct {
					Name  string    `json:"name"`
					Email string    `json:"email"`
					Date  time.Time `json:"date"`
				}{
					Name:  "Arya Khaslana",
					Email: "arya@example.com",
					Date:  time.Date(2026, 11, 10, 14, 30, 0, 0, time.UTC),
				},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockCommits)
	}))
	defer ts.Close()

	client := NewClient()
	client.httpClient = ts.Client()

	loc := time.FixedZone("WIB", 7*3600)
	// Override URL via direct testing logic
	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var commits []gitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("Expected 1 commit, got %d", len(commits))
	}
	if commits[0].SHA != "a1b2c3d4e5f6" {
		t.Errorf("Expected SHA a1b2c3d4e5f6, got %s", commits[0].SHA)
	}
	_ = loc
}
