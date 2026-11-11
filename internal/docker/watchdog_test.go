package docker

import (
	"context"
	"sync"
	"testing"
	"time"

	"devpulse/internal/domain"
	"github.com/google/uuid"
)

type mockDockerClient struct {
	containers []ContainerInfo
	err        error
	restarted  []string
	logs       string
}

func (m *mockDockerClient) ListContainers(ctx context.Context, all bool) ([]ContainerInfo, error) {
	return m.containers, m.err
}

func (m *mockDockerClient) RestartContainer(ctx context.Context, nameOrID string) error {
	m.restarted = append(m.restarted, nameOrID)
	return m.err
}

func (m *mockDockerClient) GetContainerLogs(ctx context.Context, nameOrID string, tail int) (string, error) {
	return m.logs, m.err
}

type mockAccountRepo struct {
	accounts []*domain.Account
}

func (m *mockAccountRepo) CreateOrGet(ctx context.Context, chatID int64, username string) (*domain.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) GetByChatID(ctx context.Context, chatID int64) (*domain.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) UpdateGitHub(ctx context.Context, accountID uuid.UUID, githubUsername, githubToken string) error {
	return nil
}

func (m *mockAccountRepo) UpdateStandupSchedule(ctx context.Context, accountID uuid.UUID, hour, minute int) error {
	return nil
}

func (m *mockAccountRepo) GetAll(ctx context.Context) ([]*domain.Account, error) {
	return m.accounts, nil
}

type mockNotifier struct {
	mu           sync.Mutex
	sentMessages []string
	sentKeyboards []string
}

func (m *mockNotifier) SendMessage(chatID int64, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, text)
	return nil
}

func (m *mockNotifier) SendMessageWithKeyboard(chatID int64, text string, markup interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentKeyboards = append(m.sentKeyboards, text)
	return nil
}

func TestWatchdogCrashDetection(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockDockerClient{
		containers: []ContainerInfo{
			{Names: []string{"/web-service"}, State: "running", Status: "Up 1 hour"},
			{Names: []string{"/db-service"}, State: "running", Status: "Up 2 hours"},
		},
	}
	accRepo := &mockAccountRepo{
		accounts: []*domain.Account{
			{ID: uuid.New(), TelegramChatID: 123456, Username: "testdev", CreatedAt: time.Now()},
		},
	}
	notifier := &mockNotifier{}

	watchdog := NewWatchdog(mockClient, accRepo, notifier)

	// First run: records states, no notifications sent
	watchdog.Check(ctx)
	if len(notifier.sentKeyboards) != 0 {
		t.Fatalf("Expected 0 alerts on first run, got %d", len(notifier.sentKeyboards))
	}

	// Second run: web-service is still running, db-service crashed!
	mockClient.containers = []ContainerInfo{
		{Names: []string{"/web-service"}, State: "running", Status: "Up 1 hour"},
		{Names: []string{"/db-service"}, State: "exited", Status: "Exited (137) 5 seconds ago"},
	}

	watchdog.Check(ctx)
	if len(notifier.sentKeyboards) != 1 {
		t.Fatalf("Expected 1 alert for crashed container, got %d", len(notifier.sentKeyboards))
	}

	// Verify alert content
	alert := notifier.sentKeyboards[0]
	if !contains(alert, "db-service") || !contains(alert, "MATI / CRASH") {
		t.Errorf("Alert text missing container name or crash notice: %s", alert)
	}

	// Third run: db-service is still exited, no duplicate alert should be sent
	watchdog.Check(ctx)
	if len(notifier.sentKeyboards) != 1 {
		t.Fatalf("Expected still 1 alert (no duplicate spam), got %d", len(notifier.sentKeyboards))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
