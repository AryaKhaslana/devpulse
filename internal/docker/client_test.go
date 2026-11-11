package docker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContainerInfoHelpers(t *testing.T) {
	c := &ContainerInfo{
		ID:     "1234567890abcdef",
		Names:  []string{"/pocketpulse-bot"},
		State:  "running",
		Status: "Up 2 hours",
	}

	if c.CleanName() != "pocketpulse-bot" {
		t.Errorf("Expected CleanName 'pocketpulse-bot', got %s", c.CleanName())
	}
	if !c.IsRunning() {
		t.Errorf("Expected IsRunning true")
	}

	cExited := &ContainerInfo{
		Names: []string{"/crashed-app"},
		State: "exited",
	}
	if cExited.IsRunning() {
		t.Errorf("Expected IsRunning false for exited container")
	}
}

func TestCleanDockerLogStream(t *testing.T) {
	raw := []byte("plain text log without header\nsecond line")
	cleaned := cleanDockerLogStream(raw)
	if !strings.Contains(cleaned, "plain text log") {
		t.Errorf("Expected plain text log to be preserved, got: %s", cleaned)
	}

	emptyCleaned := cleanDockerLogStream(nil)
	if emptyCleaned != "(Log kosong)" {
		t.Errorf("Expected '(Log kosong)' for empty log, got: %s", emptyCleaned)
	}
}

func TestDockerClientMock(t *testing.T) {
	mockContainers := []ContainerInfo{
		{
			ID:     "abcd1234efgh",
			Names:  []string{"/devpulse-bot"},
			State:  "running",
			Status: "Up 10 minutes",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/containers/json") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockContainers)
			return
		}
		if strings.Contains(r.URL.Path, "/restart") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if strings.Contains(r.URL.Path, "/logs") {
			w.Write([]byte("mock container log output"))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewClient("")
	client.client = ts.Client()

	// Override URL to hit mock test server
	req, _ := http.NewRequest("GET", ts.URL+"/containers/json", nil)
	resp, err := client.client.Do(req)
	if err != nil {
		t.Fatalf("Failed request to mock: %v", err)
	}
	defer resp.Body.Close()

	var containers []ContainerInfo
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if len(containers) != 1 || containers[0].CleanName() != "devpulse-bot" {
		t.Errorf("Unexpected container info: %+v", containers)
	}
}
