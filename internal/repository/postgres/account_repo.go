package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"devpulse/internal/domain"
)

type AccountRepo struct {
	pool *pgxpool.Pool
}

func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{pool: pool}
}

func (r *AccountRepo) GetByChatID(ctx context.Context, chatID int64) (*domain.Account, error) {
	query := `
		SELECT id, telegram_chat_id, username, COALESCE(github_username, ''), COALESCE(github_token, ''),
		       timezone, standup_hour, standup_minute, created_at, updated_at
		FROM accounts
		WHERE telegram_chat_id = $1
	`
	acc := &domain.Account{}
	err := r.pool.QueryRow(ctx, query, chatID).Scan(
		&acc.ID,
		&acc.TelegramChatID,
		&acc.Username,
		&acc.GitHubUsername,
		&acc.GitHubToken,
		&acc.Timezone,
		&acc.StandupHour,
		&acc.StandupMinute,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query account by chat_id: %w", err)
	}
	return acc, nil
}

func (r *AccountRepo) CreateOrGet(ctx context.Context, chatID int64, username string) (*domain.Account, error) {
	query := `
		INSERT INTO accounts (telegram_chat_id, username)
		VALUES ($1, $2)
		ON CONFLICT (telegram_chat_id) DO UPDATE
		SET username = EXCLUDED.username, updated_at = CURRENT_TIMESTAMP
		RETURNING id, telegram_chat_id, username, COALESCE(github_username, ''), COALESCE(github_token, ''),
		          timezone, standup_hour, standup_minute, created_at, updated_at
	`
	acc := &domain.Account{}
	err := r.pool.QueryRow(ctx, query, chatID, username).Scan(
		&acc.ID,
		&acc.TelegramChatID,
		&acc.Username,
		&acc.GitHubUsername,
		&acc.GitHubToken,
		&acc.Timezone,
		&acc.StandupHour,
		&acc.StandupMinute,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create or get account: %w", err)
	}
	return acc, nil
}

func (r *AccountRepo) UpdateGitHub(ctx context.Context, accountID uuid.UUID, githubUsername, githubToken string) error {
	query := `
		UPDATE accounts
		SET github_username = $2, github_token = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, accountID, githubUsername, githubToken)
	if err != nil {
		return fmt.Errorf("failed to update github profile: %w", err)
	}
	return nil
}

func (r *AccountRepo) UpdateStandupSchedule(ctx context.Context, accountID uuid.UUID, hour, minute int) error {
	query := `
		UPDATE accounts
		SET standup_hour = $2, standup_minute = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, accountID, hour, minute)
	if err != nil {
		return fmt.Errorf("failed to update standup schedule: %w", err)
	}
	return nil
}

func (r *AccountRepo) GetAll(ctx context.Context) ([]*domain.Account, error) {
	query := `
		SELECT id, telegram_chat_id, username, COALESCE(github_username, ''), COALESCE(github_token, ''),
		       timezone, standup_hour, standup_minute, created_at, updated_at
		FROM accounts
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		acc := &domain.Account{}
		if err := rows.Scan(
			&acc.ID,
			&acc.TelegramChatID,
			&acc.Username,
			&acc.GitHubUsername,
			&acc.GitHubToken,
			&acc.Timezone,
			&acc.StandupHour,
			&acc.StandupMinute,
			&acc.CreatedAt,
			&acc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}
