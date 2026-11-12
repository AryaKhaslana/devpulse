package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Idea struct {
	ID               int       `json:"id"`
	AccountID        uuid.UUID `json:"account_id"`
	Title            string    `json:"title"`
	IsArchived       bool      `json:"is_archived"`
	PromotedToTodoID *int      `json:"promoted_to_todo_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type IdeaRepository interface {
	Create(ctx context.Context, idea *Idea) error
	GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*Idea, error)
	MarkArchived(ctx context.Context, id int, accountID uuid.UUID) (*Idea, error)
	PromoteToTodo(ctx context.Context, ideaID int, accountID uuid.UUID, todoTitle string) (*Idea, *Todo, error)
	Delete(ctx context.Context, id int, accountID uuid.UUID) error
}
