package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LogCategory string

const (
	CategoryCode     LogCategory = "CODE"
	CategoryMeeting  LogCategory = "MEETING"
	CategoryBugfix   LogCategory = "BUGFIX"
	CategoryResearch LogCategory = "RESEARCH"
	CategoryDeploy   LogCategory = "DEPLOY"
	CategoryTask     LogCategory = "TASK"
	CategoryMisc     LogCategory = "MISC"
)

func (c LogCategory) Icon() string {
	switch c {
	case CategoryCode:
		return "💻"
	case CategoryMeeting:
		return "👥"
	case CategoryBugfix:
		return "🐛"
	case CategoryResearch:
		return "🔍"
	case CategoryDeploy:
		return "🚀"
	case CategoryTask:
		return "📝"
	default:
		return "📌"
	}
}

type WorkLog struct {
	ID         uuid.UUID   `json:"id"`
	AccountID  uuid.UUID   `json:"account_id"`
	Category   LogCategory `json:"category"`
	Source     string      `json:"source"` // "GIT" or "MANUAL"
	RepoName   string      `json:"repo_name,omitempty"`
	Title      string      `json:"title"`
	Details    string      `json:"details,omitempty"`
	CommitHash string      `json:"commit_hash,omitempty"`
	LogDate    time.Time   `json:"log_date"`
	CreatedAt  time.Time   `json:"created_at"`
}

type WorkLogRepository interface {
	Create(ctx context.Context, log *WorkLog) error
	CreateBatchGitCommits(ctx context.Context, logs []*WorkLog) (int, error)
	GetByDateRange(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) ([]*WorkLog, error)
	GetTodayLogs(ctx context.Context, accountID uuid.UUID, loc *time.Location) ([]*WorkLog, error)
	Search(ctx context.Context, accountID uuid.UUID, query string, limit int) ([]*WorkLog, error)
	Delete(ctx context.Context, id uuid.UUID, accountID uuid.UUID) error
}
