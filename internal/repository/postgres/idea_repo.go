package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse/internal/domain"
)

type IdeaRepo struct {
	pool *pgxpool.Pool
}

func NewIdeaRepo(pool *pgxpool.Pool) *IdeaRepo {
	return &IdeaRepo{pool: pool}
}

func (r *IdeaRepo) Create(ctx context.Context, idea *domain.Idea) error {
	query := `
		INSERT INTO ideas (account_id, title)
		VALUES ($1, $2)
		RETURNING id, is_archived, created_at
	`
	return r.pool.QueryRow(ctx, query, idea.AccountID, strings.TrimSpace(idea.Title)).
		Scan(&idea.ID, &idea.IsArchived, &idea.CreatedAt)
}

func (r *IdeaRepo) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Idea, error) {
	query := `
		SELECT id, account_id, title, is_archived, promoted_to_todo_id, created_at
		FROM ideas
		WHERE account_id = $1 AND is_archived = FALSE
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active ideas: %w", err)
	}
	defer rows.Close()

	var ideas []*domain.Idea
	for rows.Next() {
		i := &domain.Idea{}
		if err := rows.Scan(&i.ID, &i.AccountID, &i.Title, &i.IsArchived, &i.PromotedToTodoID, &i.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan idea: %w", err)
		}
		ideas = append(ideas, i)
	}
	return ideas, nil
}

func (r *IdeaRepo) MarkArchived(ctx context.Context, id int, accountID uuid.UUID) (*domain.Idea, error) {
	query := `
		UPDATE ideas
		SET is_archived = TRUE
		WHERE id = $1 AND account_id = $2
		RETURNING id, account_id, title, is_archived, promoted_to_todo_id, created_at
	`
	i := &domain.Idea{}
	err := r.pool.QueryRow(ctx, query, id, accountID).
		Scan(&i.ID, &i.AccountID, &i.Title, &i.IsArchived, &i.PromotedToTodoID, &i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to archive idea: %w", err)
	}
	return i, nil
}

func (r *IdeaRepo) PromoteToTodo(ctx context.Context, ideaID int, accountID uuid.UUID, todoTitle string) (*domain.Idea, *domain.Todo, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Fetch idea first
	var currentTitle string
	var currentArchived bool
	checkQuery := `SELECT title, is_archived FROM ideas WHERE id = $1 AND account_id = $2 FOR UPDATE`
	if err := tx.QueryRow(ctx, checkQuery, ideaID, accountID).Scan(&currentTitle, &currentArchived); err != nil {
		return nil, nil, fmt.Errorf("idea #%d not found: %w", ideaID, err)
	}

	finalTitle := currentTitle
	if strings.TrimSpace(todoTitle) != "" {
		finalTitle = strings.TrimSpace(todoTitle)
	}

	// 1. Create Todo
	todo := &domain.Todo{
		AccountID: accountID,
		Title:     finalTitle,
	}
	todoQuery := `
		INSERT INTO todos (account_id, title)
		VALUES ($1, $2)
		RETURNING id, is_completed, created_at
	`
	if err := tx.QueryRow(ctx, todoQuery, todo.AccountID, todo.Title).
		Scan(&todo.ID, &todo.IsCompleted, &todo.CreatedAt); err != nil {
		return nil, nil, fmt.Errorf("failed to create todo from idea: %w", err)
	}

	// 2. Mark Idea as archived & link promoted_to_todo_id
	idea := &domain.Idea{}
	ideaQuery := `
		UPDATE ideas
		SET is_archived = TRUE, promoted_to_todo_id = $3
		WHERE id = $1 AND account_id = $2
		RETURNING id, account_id, title, is_archived, promoted_to_todo_id, created_at
	`
	if err := tx.QueryRow(ctx, ideaQuery, ideaID, accountID, todo.ID).
		Scan(&idea.ID, &idea.AccountID, &idea.Title, &idea.IsArchived, &idea.PromotedToTodoID, &idea.CreatedAt); err != nil {
		return nil, nil, fmt.Errorf("failed to update idea promotion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit idea promotion: %w", err)
	}

	return idea, todo, nil
}

func (r *IdeaRepo) Delete(ctx context.Context, id int, accountID uuid.UUID) error {
	query := `DELETE FROM ideas WHERE id = $1 AND account_id = $2`
	_, err := r.pool.Exec(ctx, query, id, accountID)
	return err
}
