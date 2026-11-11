package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"devpulse/internal/domain"
	"devpulse/internal/github"
	"devpulse/internal/parser"
	"devpulse/internal/synthesizer"
)

type Notifier interface {
	SendMessage(chatID int64, text string) error
}

type Handler struct {
	accountRepo  domain.AccountRepository
	repoRepo     domain.MonitoredRepoRepository
	workLogRepo  domain.WorkLogRepository
	todoRepo     domain.TodoRepository
	blockerRepo  domain.BlockerRepository
	githubClient github.Client
	sender       Notifier
	loc          *time.Location
}

func NewHandler(
	accountRepo domain.AccountRepository,
	repoRepo domain.MonitoredRepoRepository,
	workLogRepo domain.WorkLogRepository,
	todoRepo domain.TodoRepository,
	blockerRepo domain.BlockerRepository,
	githubClient github.Client,
	sender Notifier,
	loc *time.Location,
) *Handler {
	return &Handler{
		accountRepo:  accountRepo,
		repoRepo:     repoRepo,
		workLogRepo:  workLogRepo,
		todoRepo:     todoRepo,
		blockerRepo:  blockerRepo,
		githubClient: githubClient,
		sender:       sender,
		loc:          loc,
	}
}

func (h *Handler) SetSender(sender Notifier) {
	h.sender = sender
}

func (h *Handler) Dispatch(ctx context.Context, msg *Message) error {
	if msg == nil || strings.TrimSpace(msg.Text) == "" {
		return nil
	}

	text := strings.TrimSpace(msg.Text)
	chatID := msg.Chat.ID
	username := msg.Chat.Username
	if username == "" && msg.From != nil {
		username = msg.From.Username
		if username == "" {
			username = msg.From.FirstName
		}
	}

	if strings.HasPrefix(text, "/") {
		return h.handleCommand(ctx, chatID, username, text)
	}

	return h.handleCasualChat(ctx, chatID, username, text)
}

func (h *Handler) handleCommand(ctx context.Context, chatID int64, username, text string) error {
	parts := strings.Fields(text)
	cmd := strings.ToLower(parts[0])
	if atIdx := strings.Index(cmd, "@"); atIdx != -1 {
		cmd = cmd[:atIdx]
	}

	switch cmd {
	case "/start":
		_, _ = h.accountRepo.CreateOrGet(ctx, chatID, username)
		return h.sender.SendMessage(chatID, RenderStart(username))
	case "/help":
		return h.sender.SendMessage(chatID, RenderStart(username))
	case "/standup":
		return h.cmdStandup(ctx, chatID, username)
	case "/logbook":
		return h.cmdLogbook(ctx, chatID, username)
	case "/today":
		return h.cmdToday(ctx, chatID)
	case "/todo":
		return h.cmdTodo(ctx, chatID, parts[1:])
	case "/blocker":
		return h.cmdBlocker(ctx, chatID, parts[1:])
	case "/github":
		return h.cmdGitHub(ctx, chatID, parts[1:])
	case "/repo":
		return h.cmdRepo(ctx, chatID, parts[1:])
	case "/sync":
		return h.cmdSync(ctx, chatID)
	case "/cari":
		if len(parts) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/cari <kata kunci>`\nContoh: `/cari cors`")
		}
		return h.cmdSearch(ctx, chatID, strings.Join(parts[1:], " "))
	case "/catat":
		if len(parts) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/catat <kegiatan>`\nContoh: `/catat meeting evaluasi sprint`")
		}
		return h.handleCasualChat(ctx, chatID, username, strings.Join(parts[1:], " "))
	default:
		return h.sender.SendMessage(chatID, "Perintah tidak dikenali. Ketik /help untuk panduan penggunaan.")
	}
}

func (h *Handler) handleCasualChat(ctx context.Context, chatID int64, username, text string) error {
	acc, err := h.accountRepo.CreateOrGet(ctx, chatID, username)
	if err != nil {
		return h.sender.SendMessage(chatID, "Gagal menginisialisasi akun.")
	}

	parsed := parser.ParseCasualInput(text)
	now := time.Now().In(h.loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, h.loc)

	switch parsed.Type {
	case parser.ParsedTypeTodo:
		todo := &domain.Todo{
			AccountID: acc.ID,
			Title:     parsed.Title,
		}
		if err := h.todoRepo.Create(ctx, todo); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menyimpan to-do.")
		}
		return h.sender.SendMessage(chatID, RenderTodoRecorded(todo))

	case parser.ParsedTypeBlocker:
		blocker := &domain.Blocker{
			AccountID:   acc.ID,
			Description: parsed.Title,
		}
		if err := h.blockerRepo.Create(ctx, blocker); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menyimpan kendala.")
		}
		return h.sender.SendMessage(chatID, RenderBlockerRecorded(blocker))

	case parser.ParsedTypeLog:
		workLog := &domain.WorkLog{
			AccountID: acc.ID,
			Category:  parsed.Category,
			Source:    "MANUAL",
			Title:     parsed.Title,
			Details:   parsed.Details,
			LogDate:   today,
			CreatedAt: now,
		}
		if err := h.workLogRepo.Create(ctx, workLog); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menyimpan catatan kerja.")
		}
		return h.sender.SendMessage(chatID, RenderLogRecorded(workLog))

	default:
		return h.sender.SendMessage(chatID,
			"💬 Catatan tidak terdeteksi sebagai aktivitas koding atau tugas.\n\n"+
				"Gunakan format kasual seperti:\n"+
				"• `meeting evaluasi mingguan`\n"+
				"• `fix bug timeout database`\n"+
				"• `todo: slicing form registrasi`\n"+
				"• `blocker: nunggu token vps`\n\n"+
				"Atau ketik `/catat <teks>` untuk mencatat manual.")
	}
}

func (h *Handler) cmdStandup(ctx context.Context, chatID int64, username string) error {
	acc, err := h.accountRepo.CreateOrGet(ctx, chatID, username)
	if err != nil {
		return h.sender.SendMessage(chatID, "Gagal memuat akun.")
	}

	now := time.Now().In(h.loc)
	daysBack := 1
	if now.Weekday() == time.Monday {
		daysBack = 3
	}
	yesterday := now.AddDate(0, 0, -daysBack)
	yStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, h.loc)
	yEnd := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, h.loc)

	yesterdayLogs, _ := h.workLogRepo.GetByDateRange(ctx, acc.ID, yStart, yEnd)
	activeTodos, _ := h.todoRepo.GetActiveByAccountID(ctx, acc.ID)
	activeBlockers, _ := h.blockerRepo.GetActiveByAccountID(ctx, acc.ID)

	standup := synthesizer.GenerateDailyStandup(acc.Username, now, yesterdayLogs, activeTodos, activeBlockers)
	return h.sender.SendMessage(chatID, standup)
}

func (h *Handler) cmdLogbook(ctx context.Context, chatID int64, username string) error {
	acc, err := h.accountRepo.CreateOrGet(ctx, chatID, username)
	if err != nil {
		return h.sender.SendMessage(chatID, "Gagal memuat akun.")
	}

	now := time.Now().In(h.loc)
	offsetToMonday := int(now.Weekday() - time.Monday)
	monday := now.AddDate(0, 0, -offsetToMonday)
	mondayStart := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, h.loc)
	fridayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, h.loc)

	logs, err := h.workLogRepo.GetByDateRange(ctx, acc.ID, mondayStart, fridayEnd)
	if err != nil {
		return h.sender.SendMessage(chatID, "Gagal mengambil data logbook.")
	}

	friday := mondayStart.AddDate(0, 0, 4)
	report := synthesizer.GenerateWeeklyLogbook(acc.Username, mondayStart, friday, logs, h.loc)
	return h.sender.SendMessage(chatID, report)
}

func (h *Handler) cmdToday(ctx context.Context, chatID int64) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	logs, err := h.workLogRepo.GetTodayLogs(ctx, acc.ID, h.loc)
	if err != nil {
		return h.sender.SendMessage(chatID, "Gagal memuat timeline hari ini.")
	}

	summary := synthesizer.GenerateDailySummary(time.Now().In(h.loc), logs)
	return h.sender.SendMessage(chatID, summary)
}

func (h *Handler) cmdTodo(ctx context.Context, chatID int64, args []string) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	if len(args) == 0 || args[0] == "list" {
		todos, err := h.todoRepo.GetActiveByAccountID(ctx, acc.ID)
		if err != nil {
			return h.sender.SendMessage(chatID, "Gagal mengambil daftar to-do.")
		}
		return h.sender.SendMessage(chatID, RenderTodoList(todos))
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "add":
		if len(args) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/todo add <rencana>`")
		}
		todo := &domain.Todo{
			AccountID: acc.ID,
			Title:     strings.Join(args[1:], " "),
		}
		if err := h.todoRepo.Create(ctx, todo); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menyimpan to-do.")
		}
		return h.sender.SendMessage(chatID, RenderTodoRecorded(todo))

	case "done":
		if len(args) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/todo done <id>` (misal: `/todo done 1`)")
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return h.sender.SendMessage(chatID, "ID to-do harus berupa angka.")
		}
		todo, err := h.todoRepo.MarkCompleted(ctx, id, acc.ID)
		if err != nil {
			return h.sender.SendMessage(chatID, fmt.Sprintf("Gagal menyelesaikan to-do #%d (mungkin tidak ditemukan).", id))
		}

		// Auto record as work log for today
		now := time.Now().In(h.loc)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, h.loc)
		_ = h.workLogRepo.Create(ctx, &domain.WorkLog{
			AccountID: acc.ID,
			Category:  domain.CategoryTask,
			Source:    "MANUAL",
			Title:     "Selesai: " + todo.Title,
			LogDate:   today,
			CreatedAt: now,
		})

		return h.sender.SendMessage(chatID, RenderTodoDone(todo))

	default:
		return h.sender.SendMessage(chatID, "Perintah to-do: `/todo list`, `/todo add <tugas>`, `/todo done <id>`")
	}
}

func (h *Handler) cmdBlocker(ctx context.Context, chatID int64, args []string) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	if len(args) == 0 || args[0] == "list" {
		blockers, err := h.blockerRepo.GetActiveByAccountID(ctx, acc.ID)
		if err != nil {
			return h.sender.SendMessage(chatID, "Gagal mengambil daftar kendala.")
		}
		return h.sender.SendMessage(chatID, RenderBlockerList(blockers))
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "add":
		if len(args) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/blocker add <kendala>`")
		}
		blocker := &domain.Blocker{
			AccountID:   acc.ID,
			Description: strings.Join(args[1:], " "),
		}
		if err := h.blockerRepo.Create(ctx, blocker); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menyimpan kendala.")
		}
		return h.sender.SendMessage(chatID, RenderBlockerRecorded(blocker))

	case "done":
		if len(args) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/blocker done <id>` (misal: `/blocker done 1`)")
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return h.sender.SendMessage(chatID, "ID kendala harus berupa angka.")
		}
		blocker, err := h.blockerRepo.MarkResolved(ctx, id, acc.ID)
		if err != nil {
			return h.sender.SendMessage(chatID, fmt.Sprintf("Gagal menyelesaikan kendala #%d.", id))
		}
		return h.sender.SendMessage(chatID, RenderBlockerDone(blocker))

	default:
		return h.sender.SendMessage(chatID, "Perintah blocker: `/blocker list`, `/blocker add <kendala>`, `/blocker done <id>`")
	}
}

func (h *Handler) cmdGitHub(ctx context.Context, chatID int64, args []string) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	if len(args) < 2 || strings.ToLower(args[0]) != "set" {
		current := "Belum diatur"
		if acc.GitHubUsername != "" {
			current = fmt.Sprintf("`%s`", acc.GitHubUsername)
		}
		return h.sender.SendMessage(chatID, fmt.Sprintf(
			"🐙 *Pengaturan Akun GitHub*\n• Username saat ini: %s\n\n"+
				"Untuk mengatur:\n`/github set <github_username> [personal_access_token]`\n\n"+
				"_Token bersifat opsional untuk public repo, wajib untuk private repo._", current))
	}

	ghUser := args[1]
	ghToken := ""
	if len(args) >= 3 {
		ghToken = args[2]
	}

	if err := h.accountRepo.UpdateGitHub(ctx, acc.ID, ghUser, ghToken); err != nil {
		return h.sender.SendMessage(chatID, "Gagal menyimpan pengaturan GitHub.")
	}

	return h.sender.SendMessage(chatID, fmt.Sprintf("✅ Akun GitHub berhasil disetel ke: `%s`", ghUser))
}

func (h *Handler) cmdRepo(ctx context.Context, chatID int64, args []string) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	if len(args) == 0 || args[0] == "list" {
		repos, err := h.repoRepo.GetByAccountID(ctx, acc.ID)
		if err != nil {
			return h.sender.SendMessage(chatID, "Gagal mengambil daftar repositori.")
		}
		return h.sender.SendMessage(chatID, RenderRepoList(repos))
	}

	subCmd := strings.ToLower(args[0])
	switch subCmd {
	case "add":
		if len(args) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/repo add <owner/repo> [branch]`\nContoh: `/repo add AryaKhaslana/PocketPulse`")
		}
		repoName := args[1]
		branch := "main"
		if len(args) >= 3 {
			branch = args[2]
		}

		repo := &domain.MonitoredRepo{
			AccountID: acc.ID,
			RepoName:  repoName,
			Branch:    branch,
		}
		if err := h.repoRepo.Add(ctx, repo); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menambahkan repositori.")
		}

		// Trigger initial sync in background
		go func() {
			since := time.Now().AddDate(0, 0, -7)
			commits, err := h.githubClient.FetchRecentCommits(context.Background(), repoName, branch, acc.GitHubUsername, acc.GitHubToken, since, h.loc)
			if err == nil {
				for _, c := range commits {
					c.AccountID = acc.ID
				}
				ins, _ := h.workLogRepo.CreateBatchGitCommits(context.Background(), commits)
				if ins > 0 {
					_ = h.sender.SendMessage(chatID, fmt.Sprintf("⚡ Berhasil menarik %d commit awal dari `%s`!", ins, repoName))
				}
				_ = h.repoRepo.UpdateLastSynced(context.Background(), repo.ID, time.Now())
			}
		}()

		return h.sender.SendMessage(chatID, fmt.Sprintf("✅ Repositori `%s` (branch: %s) berhasil dipantau!", repoName, branch))

	case "del", "remove", "delete":
		if len(args) < 2 {
			return h.sender.SendMessage(chatID, "Format: `/repo del <owner/repo>`")
		}
		repoName := args[1]
		if err := h.repoRepo.Delete(ctx, acc.ID, repoName); err != nil {
			return h.sender.SendMessage(chatID, "Gagal menghapus repositori.")
		}
		return h.sender.SendMessage(chatID, fmt.Sprintf("🗑️ Repositori `%s` dihapus dari pantauan.", repoName))

	default:
		return h.sender.SendMessage(chatID, "Perintah repo: `/repo list`, `/repo add <owner/repo>`, `/repo del <owner/repo>`")
	}
}

func (h *Handler) cmdSync(ctx context.Context, chatID int64) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	repos, err := h.repoRepo.GetByAccountID(ctx, acc.ID)
	if err != nil || len(repos) == 0 {
		return h.sender.SendMessage(chatID, "Belum ada repositori yang dipantau. Tambah via `/repo add <owner/repo>`.")
	}

	_ = h.sender.SendMessage(chatID, "⏳ Sedang menyinkronkan aktivitas commit GitHub...")

	totalInserted := 0
	for _, r := range repos {
		var since time.Time
		if r.LastSyncedAt != nil {
			since = *r.LastSyncedAt
		} else {
			since = time.Now().AddDate(0, 0, -7)
		}

		commits, err := h.githubClient.FetchRecentCommits(ctx, r.RepoName, r.Branch, acc.GitHubUsername, acc.GitHubToken, since, h.loc)
		if err != nil {
			log.Printf("[Manual Sync] Error on %s: %v", r.RepoName, err)
			continue
		}

		for _, c := range commits {
			c.AccountID = acc.ID
		}

		inserted, _ := h.workLogRepo.CreateBatchGitCommits(ctx, commits)
		totalInserted += inserted
		_ = h.repoRepo.UpdateLastSynced(ctx, r.ID, time.Now())
	}

	return h.sender.SendMessage(chatID, fmt.Sprintf("✅ Sinkronisasi selesai! %d commit baru berhasil dicatat ke DevPulse.", totalInserted))
}

func (h *Handler) cmdSearch(ctx context.Context, chatID int64, query string) error {
	acc, err := h.accountRepo.GetByChatID(ctx, chatID)
	if err != nil || acc == nil {
		return h.sender.SendMessage(chatID, "Akun belum terdaftar. Ketik /start.")
	}

	logs, err := h.workLogRepo.Search(ctx, acc.ID, query, 10)
	if err != nil {
		return h.sender.SendMessage(chatID, "Gagal melakukan pencarian.")
	}

	return h.sender.SendMessage(chatID, RenderSearchResults(query, logs))
}
