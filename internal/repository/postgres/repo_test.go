package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"devpulse/internal/database"
	"devpulse/internal/domain"
)

func TestPostgresRepositoriesIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devpulse:devpulse_secret@localhost:5433/devpulse?sslmode=disable"
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("Postgres DB not accessible at %s, skipping integration test: %v", dbURL, err)
		return
	}
	defer db.Close()

	loc := time.FixedZone("WIB", 7*3600)
	accountRepo := NewAccountRepo(db.Pool)
	repoRepo := NewMonitoredRepoRepo(db.Pool)
	workLogRepo := NewWorkLogRepo(db.Pool)
	todoRepo := NewTodoRepo(db.Pool)
	blockerRepo := NewBlockerRepo(db.Pool)

	// Clean up previous test account if any
	testChatID := int64(999888777)
	_, _ = db.Pool.Exec(ctx, "DELETE FROM accounts WHERE telegram_chat_id = $1", testChatID)

	// 1. Account Repo Test
	acc, err := accountRepo.CreateOrGet(ctx, testChatID, "AryaDevTest")
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}
	if acc.Username != "AryaDevTest" {
		t.Errorf("Expected username AryaDevTest, got %s", acc.Username)
	}

	err = accountRepo.UpdateGitHub(ctx, acc.ID, "AryaKhaslana", "ghp_mocktoken")
	if err != nil {
		t.Fatalf("Failed to update github profile: %v", err)
	}

	accFetched, err := accountRepo.GetByChatID(ctx, testChatID)
	if err != nil || accFetched == nil {
		t.Fatalf("Failed to get account by chatID: %v", err)
	}
	if accFetched.GitHubUsername != "AryaKhaslana" {
		t.Errorf("Expected GitHubUsername AryaKhaslana, got %s", accFetched.GitHubUsername)
	}

	// 2. Monitored Repo Test
	mRepo := &domain.MonitoredRepo{
		AccountID: acc.ID,
		RepoName:  "AryaKhaslana/PocketPulse",
		Branch:    "main",
	}
	err = repoRepo.Add(ctx, mRepo)
	if err != nil {
		t.Fatalf("Failed to add monitored repo: %v", err)
	}

	repos, err := repoRepo.GetByAccountID(ctx, acc.ID)
	if err != nil || len(repos) != 1 {
		t.Fatalf("Expected 1 monitored repo, got %d (err: %v)", len(repos), err)
	}

	// 3. WorkLog Repo Test
	today := time.Date(2026, 11, 11, 0, 0, 0, 0, loc)
	now := time.Now().In(loc)

	log1 := &domain.WorkLog{
		AccountID: acc.ID,
		Category:  domain.CategoryMeeting,
		Source:    "MANUAL",
		Title:     "Briefing sprint PKL dengan mentor",
		Details:   "Bahas arsitektur database",
		LogDate:   today,
		CreatedAt: now,
	}
	err = workLogRepo.Create(ctx, log1)
	if err != nil {
		t.Fatalf("Failed to create work log: %v", err)
	}

	// Batch Git Commits Ingestion & Idempotency
	gitCommits := []*domain.WorkLog{
		{
			AccountID:  acc.ID,
			Category:   domain.CategoryCode,
			Source:     "GIT",
			RepoName:   "AryaKhaslana/PocketPulse",
			Title:      "feat: add scheduler",
			CommitHash: "sha_commit_12345",
			LogDate:    today,
			CreatedAt:  now,
		},
		{
			AccountID:  acc.ID,
			Category:   domain.CategoryCode,
			Source:     "GIT",
			RepoName:   "AryaKhaslana/PocketPulse",
			Title:      "feat: add scheduler",
			CommitHash: "sha_commit_12345", // duplicate commit
			LogDate:    today,
			CreatedAt:  now,
		},
	}
	inserted, err := workLogRepo.CreateBatchGitCommits(ctx, gitCommits)
	if err != nil {
		t.Fatalf("Failed batch git commits: %v", err)
	}
	if inserted != 1 {
		t.Errorf("Expected 1 inserted commit (1 duplicate skipped), got %d", inserted)
	}

	// Test GetTodayLogs
	todayLogs, err := workLogRepo.GetTodayLogs(ctx, acc.ID, loc)
	if err != nil {
		t.Fatalf("Failed to get today logs: %v", err)
	}
	if len(todayLogs) < 2 {
		t.Errorf("Expected at least 2 today logs, got %d", len(todayLogs))
	}

	// Test Search
	searchRes, err := workLogRepo.Search(ctx, acc.ID, "scheduler", 10)
	if err != nil {
		t.Fatalf("Failed to search work logs: %v", err)
	}
	if len(searchRes) != 1 {
		t.Errorf("Expected 1 search result for 'scheduler', got %d", len(searchRes))
	}

	// 4. Todo Repo Test
	todo := &domain.Todo{
		AccountID: acc.ID,
		Title:     "Setup CI/CD GitHub Actions",
	}
	err = todoRepo.Create(ctx, todo)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	activeTodos, err := todoRepo.GetActiveByAccountID(ctx, acc.ID)
	if err != nil || len(activeTodos) != 1 {
		t.Fatalf("Expected 1 active todo, got %d", len(activeTodos))
	}

	doneTodo, err := todoRepo.MarkCompleted(ctx, todo.ID, acc.ID)
	if err != nil {
		t.Fatalf("Failed to mark todo completed: %v", err)
	}
	if !doneTodo.IsCompleted {
		t.Errorf("Expected todo to be completed")
	}

	// 5. Blocker Repo Test
	blocker := &domain.Blocker{
		AccountID:   acc.ID,
		Description: "Akses cloudflare API belum di-whitelist",
	}
	err = blockerRepo.Create(ctx, blocker)
	if err != nil {
		t.Fatalf("Failed to create blocker: %v", err)
	}

	activeBlockers, err := blockerRepo.GetActiveByAccountID(ctx, acc.ID)
	if err != nil || len(activeBlockers) != 1 {
		t.Fatalf("Expected 1 active blocker, got %d", len(activeBlockers))
	}

	resolvedBlocker, err := blockerRepo.MarkResolved(ctx, blocker.ID, acc.ID)
	if err != nil {
		t.Fatalf("Failed to mark blocker resolved: %v", err)
	}
	if !resolvedBlocker.IsResolved {
		t.Errorf("Expected blocker to be resolved")
	}

	// Cleanup
	_, _ = db.Pool.Exec(ctx, "DELETE FROM accounts WHERE id = $1", acc.ID)
}
