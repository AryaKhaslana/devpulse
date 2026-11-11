package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse/internal/domain"
)

type BlockerRepo struct {
	pool *pgxpool.Pool
}

func NewBlockerRepo(pool *pgxpool.Pool) *BlockerRepo {
	return &BlockerRepo{pool: pool}
}

func (r *BlockerRepo) Create(ctx context.Context, blocker *domain.Blocker) error {
	query := `
		INSERT INTO blockers (account_id, description)
		VALUES ($1, $2)
		RETURNING id, is_resolved, created_at
	`
	return r.pool.QueryRow(ctx, query, blocker.AccountID, blocker.Description).Scan(&blocker.ID, &blocker.IsResolved, &blocker.CreatedAt)
}

func (r *BlockerRepo) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Blocker, error) {
	query := `
		SELECT id, account_id, description, is_resolved, resolved_at, created_at
		FROM blockers
		WHERE account_id = $1 AND is_resolved = FALSE
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active blockers: %w", err)
	}
	defer rows.Close()

	var blockers []*domain.Blocker
	for rows.Next() {
		b := &domain.Blocker{}
		if err := rows.Scan(&b.ID, &b.AccountID, &b.Description, &b.IsResolved, &b.ResolvedAt, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan blocker: %w", err)
		}
		blockers = append(blockers, b)
	}
	return blockers, nil
}

func (r *BlockerRepo) MarkResolved(ctx context.Context, id int, accountID uuid.UUID) (*domain.Blocker, error) {
	now := time.Now()
	query := `
		UPDATE blockers
		SET is_resolved = TRUE, resolved_at = $3
		WHERE id = $1 AND account_id = $2
		RETURNING id, account_id, description, is_resolved, resolved_at, created_at
	`
	b := &domain.Blocker{}
	err := r.pool.QueryRow(ctx, query, id, accountID, now).Scan(&b.ID, &b.AccountID, &b.Description, &b.IsResolved, &b.ResolvedAt, &b.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to mark blocker resolved: %w", err)
	}
	return b, nil
}

func (r *BlockerRepo) Delete(ctx context.Context, id int, accountID uuid.UUID) error {
	query := `DELETE FROM blockers WHERE id = $1 AND account_id = $2`
	_, err := r.pool.Exec(ctx, query, id, accountID)
	return err
}
