package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Blocker struct {
	ID          int        `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Description string     `json:"description"`
	IsResolved  bool       `json:"is_resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type BlockerRepository interface {
	Create(ctx context.Context, blocker *Blocker) error
	GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*Blocker, error)
	MarkResolved(ctx context.Context, id int, accountID uuid.UUID) (*Blocker, error)
	Delete(ctx context.Context, id int, accountID uuid.UUID) error
}
