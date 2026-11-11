package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MonitoredRepo struct {
	ID           int        `json:"id"`
	AccountID    uuid.UUID  `json:"account_id"`
	RepoName     string     `json:"repo_name"` // "owner/repo"
	Branch       string     `json:"branch"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type MonitoredRepoRepository interface {
	Add(ctx context.Context, repo *MonitoredRepo) error
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*MonitoredRepo, error)
	Delete(ctx context.Context, accountID uuid.UUID, repoName string) error
	UpdateLastSynced(ctx context.Context, id int, syncedAt time.Time) error
}
