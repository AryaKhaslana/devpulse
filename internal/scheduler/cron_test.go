package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"devpulse/internal/domain"
)

type mockAccountRepo struct {
	acc *domain.Account
}

func (m *mockAccountRepo) GetByChatID(ctx context.Context, chatID int64) (*domain.Account, error) {
	return m.acc, nil
}
func (m *mockAccountRepo) CreateOrGet(ctx context.Context, chatID int64, username string) (*domain.Account, error) {
	return m.acc, nil
}
func (m *mockAccountRepo) UpdateGitHub(ctx context.Context, accountID uuid.UUID, githubUsername, githubToken string) error {
	return nil
}
func (m *mockAccountRepo) UpdateStandupSchedule(ctx context.Context, accountID uuid.UUID, hour, minute int) error {
	return nil
}
func (m *mockAccountRepo) GetAll(ctx context.Context) ([]*domain.Account, error) {
	return []*domain.Account{m.acc}, nil
}

type mockNotifier struct {
	lastChatID int64
	lastMsg    string
}

func (m *mockNotifier) SendMessage(chatID int64, text string) error {
	m.lastChatID = chatID
	m.lastMsg = text
	return nil
}

type mockWorkLogRepo struct{}

func (m *mockWorkLogRepo) Create(ctx context.Context, log *domain.WorkLog) error { return nil }
func (m *mockWorkLogRepo) CreateBatchGitCommits(ctx context.Context, logs []*domain.WorkLog) (int, error) {
	return len(logs), nil
}
func (m *mockWorkLogRepo) GetByDateRange(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) ([]*domain.WorkLog, error) {
	return []*domain.WorkLog{
		{
			Category: domain.CategoryCode,
			RepoName: "devpulse",
			Title:    "feat: scheduler implementation",
		},
	}, nil
}
func (m *mockWorkLogRepo) GetTodayLogs(ctx context.Context, accountID uuid.UUID, loc *time.Location) ([]*domain.WorkLog, error) {
	return nil, nil
}
func (m *mockWorkLogRepo) Search(ctx context.Context, accountID uuid.UUID, query string, limit int) ([]*domain.WorkLog, error) {
	return nil, nil
}
func (m *mockWorkLogRepo) Delete(ctx context.Context, id uuid.UUID, accountID uuid.UUID) error {
	return nil
}

type mockTodoRepo struct{}

func (m *mockTodoRepo) Create(ctx context.Context, todo *domain.Todo) error { return nil }
func (m *mockTodoRepo) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Todo, error) {
	return []*domain.Todo{{Title: "Test todo"}}, nil
}
func (m *mockTodoRepo) MarkCompleted(ctx context.Context, id int, accountID uuid.UUID) (*domain.Todo, error) {
	return nil, nil
}
func (m *mockTodoRepo) Delete(ctx context.Context, id int, accountID uuid.UUID) error { return nil }

type mockBlockerRepo struct{}

func (m *mockBlockerRepo) Create(ctx context.Context, blocker *domain.Blocker) error { return nil }
func (m *mockBlockerRepo) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Blocker, error) {
	return []*domain.Blocker{}, nil
}
func (m *mockBlockerRepo) MarkResolved(ctx context.Context, id int, accountID uuid.UUID) (*domain.Blocker, error) {
	return nil, nil
}
func (m *mockBlockerRepo) Delete(ctx context.Context, id int, accountID uuid.UUID) error { return nil }

func TestDailyStandupJob(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)
	notifier := &mockNotifier{}
	acc := &domain.Account{
		ID:             uuid.New(),
		TelegramChatID: 123456,
		Username:       "AryaKhaslana",
	}

	sched := NewScheduler(
		&mockAccountRepo{acc: acc},
		nil,
		&mockWorkLogRepo{},
		&mockTodoRepo{},
		&mockBlockerRepo{},
		nil,
		notifier,
		loc,
	)

	sched.RunDailyStandupJob(context.Background())

	if notifier.lastChatID != 123456 {
		t.Errorf("Expected chat ID 123456, got %d", notifier.lastChatID)
	}
	if notifier.lastMsg == "" {
		t.Errorf("Expected standup message sent, got empty string")
	}
}
