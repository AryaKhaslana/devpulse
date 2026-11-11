package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse/internal/domain"
)

type TodoRepo struct {
	pool *pgxpool.Pool
}

func NewTodoRepo(pool *pgxpool.Pool) *TodoRepo {
	return &TodoRepo{pool: pool}
}

func (r *TodoRepo) Create(ctx context.Context, todo *domain.Todo) error {
	query := `
		INSERT INTO todos (account_id, title)
		VALUES ($1, $2)
		RETURNING id, is_completed, created_at
	`
	return r.pool.QueryRow(ctx, query, todo.AccountID, todo.Title).Scan(&todo.ID, &todo.IsCompleted, &todo.CreatedAt)
}

func (r *TodoRepo) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Todo, error) {
	query := `
		SELECT id, account_id, title, is_completed, completed_at, created_at
		FROM todos
		WHERE account_id = $1 AND is_completed = FALSE
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active todos: %w", err)
	}
	defer rows.Close()

	var todos []*domain.Todo
	for rows.Next() {
		t := &domain.Todo{}
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Title, &t.IsCompleted, &t.CompletedAt, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *TodoRepo) MarkCompleted(ctx context.Context, id int, accountID uuid.UUID) (*domain.Todo, error) {
	now := time.Now()
	query := `
		UPDATE todos
		SET is_completed = TRUE, completed_at = $3
		WHERE id = $1 AND account_id = $2
		RETURNING id, account_id, title, is_completed, completed_at, created_at
	`
	t := &domain.Todo{}
	err := r.pool.QueryRow(ctx, query, id, accountID, now).Scan(&t.ID, &t.AccountID, &t.Title, &t.IsCompleted, &t.CompletedAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to mark todo completed: %w", err)
	}
	return t, nil
}

func (r *TodoRepo) Delete(ctx context.Context, id int, accountID uuid.UUID) error {
	query := `DELETE FROM todos WHERE id = $1 AND account_id = $2`
	_, err := r.pool.Exec(ctx, query, id, accountID)
	return err
}
