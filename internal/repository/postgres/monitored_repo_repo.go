package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse/internal/domain"
)

type MonitoredRepoRepo struct {
	pool *pgxpool.Pool
}

func NewMonitoredRepoRepo(pool *pgxpool.Pool) *MonitoredRepoRepo {
	return &MonitoredRepoRepo{pool: pool}
}

func (r *MonitoredRepoRepo) Add(ctx context.Context, repo *domain.MonitoredRepo) error {
	query := `
		INSERT INTO monitored_repos (account_id, repo_name, branch)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id, repo_name) DO UPDATE
		SET branch = EXCLUDED.branch
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, repo.AccountID, repo.RepoName, repo.Branch).Scan(&repo.ID, &repo.CreatedAt)
}

func (r *MonitoredRepoRepo) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.MonitoredRepo, error) {
	query := `
		SELECT id, account_id, repo_name, branch, last_synced_at, created_at
		FROM monitored_repos
		WHERE account_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query monitored repos: %w", err)
	}
	defer rows.Close()

	var list []*domain.MonitoredRepo
	for rows.Next() {
		m := &domain.MonitoredRepo{}
		if err := rows.Scan(&m.ID, &m.AccountID, &m.RepoName, &m.Branch, &m.LastSyncedAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan monitored repo: %w", err)
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *MonitoredRepoRepo) Delete(ctx context.Context, accountID uuid.UUID, repoName string) error {
	query := `DELETE FROM monitored_repos WHERE account_id = $1 AND repo_name = $2`
	_, err := r.pool.Exec(ctx, query, accountID, repoName)
	if err != nil {
		return fmt.Errorf("failed to delete monitored repo: %w", err)
	}
	return nil
}

func (r *MonitoredRepoRepo) UpdateLastSynced(ctx context.Context, id int, syncedAt time.Time) error {
	query := `UPDATE monitored_repos SET last_synced_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, syncedAt)
	return err
}
