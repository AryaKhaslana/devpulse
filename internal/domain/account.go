package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             uuid.UUID `json:"id"`
	TelegramChatID int64     `json:"telegram_chat_id"`
	Username       string    `json:"username"`
	GitHubUsername string    `json:"github_username"`
	GitHubToken    string    `json:"-"`
	Timezone       string    `json:"timezone"`
	StandupHour    int       `json:"standup_hour"`
	StandupMinute  int       `json:"standup_minute"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AccountRepository interface {
	GetByChatID(ctx context.Context, chatID int64) (*Account, error)
	CreateOrGet(ctx context.Context, chatID int64, username string) (*Account, error)
	UpdateGitHub(ctx context.Context, accountID uuid.UUID, githubUsername, githubToken string) error
	UpdateStandupSchedule(ctx context.Context, accountID uuid.UUID, hour, minute int) error
	GetAll(ctx context.Context) ([]*Account, error)
}
