package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID          int        `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Title       string     `json:"title"`
	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*Todo, error)
	MarkCompleted(ctx context.Context, id int, accountID uuid.UUID) (*Todo, error)
	Delete(ctx context.Context, id int, accountID uuid.UUID) error
}
