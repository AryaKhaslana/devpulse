package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"devpulse/internal/domain"
	"devpulse/internal/github"
	"devpulse/internal/synthesizer"
)

type Notifier interface {
	SendMessage(chatID int64, text string) error
}

type Scheduler struct {
	cron         *cron.Cron
	accountRepo  domain.AccountRepository
	repoRepo     domain.MonitoredRepoRepository
	workLogRepo  domain.WorkLogRepository
	todoRepo     domain.TodoRepository
	blockerRepo  domain.BlockerRepository
	githubClient github.Client
	notifier     Notifier
	loc          *time.Location
}

func NewScheduler(
	accountRepo domain.AccountRepository,
	repoRepo domain.MonitoredRepoRepository,
	workLogRepo domain.WorkLogRepository,
	todoRepo domain.TodoRepository,
	blockerRepo domain.BlockerRepository,
	githubClient github.Client,
	notifier Notifier,
	loc *time.Location,
) *Scheduler {
	c := cron.New(cron.WithLocation(loc))
	return &Scheduler{
		cron:         c,
		accountRepo:  accountRepo,
		repoRepo:     repoRepo,
		workLogRepo:  workLogRepo,
		todoRepo:     todoRepo,
		blockerRepo:  blockerRepo,
		githubClient: githubClient,
		notifier:     notifier,
		loc:          loc,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	// 1. Morning Standup: Monday to Friday at 08:30 WIB
	_, err := s.cron.AddFunc("30 8 * * 1-5", func() {
		s.RunDailyStandupJob(ctx)
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule morning standup: %v", err)
	}

	// 2. Weekly Logbook: Friday at 17:00 WIB
	_, err = s.cron.AddFunc("0 17 * * 5", func() {
		s.RunWeeklyLogbookJob(ctx)
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule weekly logbook: %v", err)
	}

	// 3. Periodic GitHub Commit Sync: every 30 minutes
	_, err = s.cron.AddFunc("*/30 * * * *", func() {
		s.RunGitHubSyncJob(ctx)
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule git sync: %v", err)
	}

	s.cron.Start()
	log.Printf("[Scheduler] Automated standup (08:30), logbook (Jum 17:00), & git sync cron started (%s)", s.loc.String())
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *Scheduler) RunDailyStandupJob(ctx context.Context) {
	now := time.Now().In(s.loc)
	log.Printf("[Scheduler] Executing Daily Standup job at %s", now.Format(time.RFC3339))

	accounts, err := s.accountRepo.GetAll(ctx)
	if err != nil {
		log.Printf("[Scheduler] Failed to get accounts for standup: %v", err)
		return
	}

	// Determine "yesterday" (if Monday, yesterday is Friday: -3 days; otherwise -1 day)
	daysBack := 1
	if now.Weekday() == time.Monday {
		daysBack = 3
	}
	yesterday := now.AddDate(0, 0, -daysBack)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, s.loc)
	yesterdayEnd := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, s.loc)

	for _, acc := range accounts {
		logs, err := s.workLogRepo.GetByDateRange(ctx, acc.ID, yesterdayStart, yesterdayEnd)
		if err != nil {
			log.Printf("[Scheduler] Failed to get yesterday logs for user %d: %v", acc.TelegramChatID, err)
			continue
		}

		todos, _ := s.todoRepo.GetActiveByAccountID(ctx, acc.ID)
		blockers, _ := s.blockerRepo.GetActiveByAccountID(ctx, acc.ID)

		standupText := synthesizer.GenerateDailyStandup(acc.Username, now, logs, todos, blockers)
		msg := fmt.Sprintf("☀️ *Selamat Pagi! Draf Daily Standup Anda Hari Ini Siap:*\n\n%s\n\n💡 _Tip: Salin dan bagikan ke tim Anda._", standupText)

		if s.notifier != nil {
			_ = s.notifier.SendMessage(acc.TelegramChatID, msg)
		}
	}
}

func (s *Scheduler) RunWeeklyLogbookJob(ctx context.Context) {
	now := time.Now().In(s.loc)
	log.Printf("[Scheduler] Executing Weekly Logbook job at %s", now.Format(time.RFC3339))

	accounts, err := s.accountRepo.GetAll(ctx)
	if err != nil {
		log.Printf("[Scheduler] Failed to get accounts for logbook: %v", err)
		return
	}

	// Monday of current week
	offsetToMonday := int(now.Weekday() - time.Monday)
	monday := now.AddDate(0, 0, -offsetToMonday)
	mondayStart := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, s.loc)
	fridayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, s.loc)

	for _, acc := range accounts {
		logs, err := s.workLogRepo.GetByDateRange(ctx, acc.ID, mondayStart, fridayEnd)
		if err != nil {
			log.Printf("[Scheduler] Failed to get week logs for user %d: %v", acc.TelegramChatID, err)
			continue
		}

		friday := mondayStart.AddDate(0, 0, 4)
		logbookText := synthesizer.GenerateWeeklyLogbook(acc.Username, mondayStart, friday, logs, s.loc)
		msg := fmt.Sprintf("🎉 *Happy Friday! Rekap Laporan Mingguan PKL/Kerja:*\n\n%s", logbookText)

		if s.notifier != nil {
			_ = s.notifier.SendMessage(acc.TelegramChatID, msg)
		}
	}
}

func (s *Scheduler) RunGitHubSyncJob(ctx context.Context) {
	accounts, err := s.accountRepo.GetAll(ctx)
	if err != nil {
		return
	}

	for _, acc := range accounts {
		repos, err := s.repoRepo.GetByAccountID(ctx, acc.ID)
		if err != nil || len(repos) == 0 {
			continue
		}

		for _, r := range repos {
			var since time.Time
			if r.LastSyncedAt != nil {
				since = *r.LastSyncedAt
			} else {
				// Default sync up to 7 days back
				since = time.Now().AddDate(0, 0, -7)
			}

			commits, err := s.githubClient.FetchRecentCommits(ctx, r.RepoName, r.Branch, acc.GitHubUsername, acc.GitHubToken, since, s.loc)
			if err != nil {
				log.Printf("[Scheduler GitSync] Error syncing %s: %v", r.RepoName, err)
				continue
			}

			for _, c := range commits {
				c.AccountID = acc.ID
			}

			inserted, _ := s.workLogRepo.CreateBatchGitCommits(ctx, commits)
			if inserted > 0 {
				log.Printf("[Scheduler GitSync] Synced %d new commits for user %d from %s", inserted, acc.TelegramChatID, r.RepoName)
			}
			_ = s.repoRepo.UpdateLastSynced(ctx, r.ID, time.Now())
		}
	}
}
