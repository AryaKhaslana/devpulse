package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse/internal/domain"
)

type WorkLogRepo struct {
	pool *pgxpool.Pool
}

func NewWorkLogRepo(pool *pgxpool.Pool) *WorkLogRepo {
	return &WorkLogRepo{pool: pool}
}

func (r *WorkLogRepo) Create(ctx context.Context, log *domain.WorkLog) error {
	query := `
		INSERT INTO work_logs (account_id, category, source, repo_name, title, details, commit_hash, log_date, created_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), NULLIF($7, ''), $8, COALESCE($9, CURRENT_TIMESTAMP))
		RETURNING id, created_at
	`
	var createdAt *time.Time
	if !log.CreatedAt.IsZero() {
		createdAt = &log.CreatedAt
	}

	return r.pool.QueryRow(
		ctx, query,
		log.AccountID,
		log.Category,
		log.Source,
		log.RepoName,
		log.Title,
		log.Details,
		log.CommitHash,
		log.LogDate,
		createdAt,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *WorkLogRepo) CreateBatchGitCommits(ctx context.Context, logs []*domain.WorkLog) (int, error) {
	if len(logs) == 0 {
		return 0, nil
	}

	insertedCount := 0
	query := `
		INSERT INTO work_logs (account_id, category, source, repo_name, title, details, commit_hash, log_date, created_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9)
		ON CONFLICT (account_id, commit_hash) WHERE commit_hash IS NOT NULL AND commit_hash != '' DO NOTHING
		RETURNING id
	`

	for _, l := range logs {
		var id uuid.UUID
		err := r.pool.QueryRow(
			ctx, query,
			l.AccountID,
			l.Category,
			l.Source,
			l.RepoName,
			l.Title,
			l.Details,
			l.CommitHash,
			l.LogDate,
			l.CreatedAt,
		).Scan(&id)
		if err == nil {
			insertedCount++
		}
	}

	return insertedCount, nil
}

func (r *WorkLogRepo) GetByDateRange(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) ([]*domain.WorkLog, error) {
	query := `
		SELECT id, account_id, category, source, COALESCE(repo_name, ''),
		       title, COALESCE(details, ''), COALESCE(commit_hash, ''),
		       log_date, created_at
		FROM work_logs
		WHERE account_id = $1 AND log_date >= $2 AND log_date <= $3
		ORDER BY log_date ASC, created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, accountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query work logs by date range: %w", err)
	}
	defer rows.Close()

	var logs []*domain.WorkLog
	for rows.Next() {
		l := &domain.WorkLog{}
		if err := rows.Scan(
			&l.ID,
			&l.AccountID,
			&l.Category,
			&l.Source,
			&l.RepoName,
			&l.Title,
			&l.Details,
			&l.CommitHash,
			&l.LogDate,
			&l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan work log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *WorkLogRepo) GetTodayLogs(ctx context.Context, accountID uuid.UUID, loc *time.Location) ([]*domain.WorkLog, error) {
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return r.GetByDateRange(ctx, accountID, today, today)
}

func (r *WorkLogRepo) Search(ctx context.Context, accountID uuid.UUID, queryStr string, limit int) ([]*domain.WorkLog, error) {
	if limit <= 0 {
		limit = 10
	}
	likePattern := "%" + queryStr + "%"
	query := `
		SELECT id, account_id, category, source, COALESCE(repo_name, ''),
		       title, COALESCE(details, ''), COALESCE(commit_hash, ''),
		       log_date, created_at
		FROM work_logs
		WHERE account_id = $1 AND (title ILIKE $2 OR details ILIKE $2 OR repo_name ILIKE $2)
		ORDER BY log_date DESC, created_at DESC
		LIMIT $3
	`
	rows, err := r.pool.Query(ctx, query, accountID, likePattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search work logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.WorkLog
	for rows.Next() {
		l := &domain.WorkLog{}
		if err := rows.Scan(
			&l.ID,
			&l.AccountID,
			&l.Category,
			&l.Source,
			&l.RepoName,
			&l.Title,
			&l.Details,
			&l.CommitHash,
			&l.LogDate,
			&l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan search work log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *WorkLogRepo) Delete(ctx context.Context, id uuid.UUID, accountID uuid.UUID) error {
	query := `DELETE FROM work_logs WHERE id = $1 AND account_id = $2`
	_, err := r.pool.Exec(ctx, query, id, accountID)
	return err
}
