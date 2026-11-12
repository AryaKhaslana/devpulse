package telegram

import (
	"fmt"
	"strings"

	"devpulse/internal/docker"
	"devpulse/internal/domain"
	"devpulse/internal/synthesizer"
)

func RenderStart(username string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👋 Halo, %s! Selamat datang di DevPulse.\n", username))
	sb.WriteString("Asisten pintar pencatat aktivitas koding, daily standup, & logbook PKL otomatis.\n\n")

	sb.WriteString("⚡ Cara Pakai Cepat (Ketik Santai):\n")
	sb.WriteString("• Koding: _ngoding endpoint settlement piutang_\n")
	sb.WriteString("• Meeting: _meeting sama pak andi bahas arsitektur database_\n")
	sb.WriteString("• Bugfix: _fix bug cors reverse proxy satpam-nginx_\n")
	sb.WriteString("• Riset: _riset redis queue vs rabbitmq buat pkl_\n")
	sb.WriteString("• Deploy: _deploy staging docker-compose di vps_\n")
	sb.WriteString("• Rencana: _todo: bikin unit test endpoint login_\n")
	sb.WriteString("• Kendala: _blocker: nunggu token cloudflare dari mentor_\n\n")

	sb.WriteString("📋 Perintah Tersedia:\n")
	sb.WriteString("• /standup - Buat draf Daily Standup hari ini\n")
	sb.WriteString("• /logbook - Generate Laporan Mingguan PKL (Senin-Jumat)\n")
	sb.WriteString("• /today - Timeline seluruh aktivitas hari ini\n")
	sb.WriteString("• /todo - Kelola to-do list aktif\n")
	sb.WriteString("• /blocker - Kelola kendala / blocker aktif\n")
	sb.WriteString("• /github set <user> [token] - Hubungkan akun GitHub\n")
	sb.WriteString("• /repo add <owner/repo> - Pantau commit GitHub\n")
	sb.WriteString("• /repo list - Daftar repo yang dipantau\n")
	sb.WriteString("• /sync - Tarik commit terbaru GitHub sekarang\n")
	sb.WriteString("• /cari <kata kunci> - Cari riwayat catatan masa lalu\n")
	sb.WriteString("• /containers atau /docker - Pantau status & restart Docker container\n")
	sb.WriteString("• /ide - Idea Vault & Brain Dump (simpan ide tanpa mengotori to-do)\n")
	return sb.String()
}

func RenderLogRecorded(log *domain.WorkLog) string {
	repoInfo := ""
	if log.RepoName != "" {
		repoInfo = fmt.Sprintf("\n• Repo: `%s`", log.RepoName)
	}
	return fmt.Sprintf("📝 *Aktivitas Tercatat!*\n• Kategori: %s %s%s\n• Catatan: %s\n• Tanggal: %s",
		log.Category.Icon(), log.Category, repoInfo, log.Title, synthesizer.FormatIndonesianDate(log.LogDate))
}

func RenderTodoRecorded(todo *domain.Todo) string {
	return fmt.Sprintf("🎯 *To-Do Ditambahkan!*\n• ID: #%d\n• Rencana: %s\n\nℹ️ Ketik `/todo done %d` jika sudah selesai.", todo.ID, todo.Title, todo.ID)
}

func RenderTodoList(todos []*domain.Todo) string {
	if len(todos) == 0 {
		return "✨ Hebat! Tidak ada to-do aktif. Tambah via `todo: <rencana>` atau `/todo add <rencana>`."
	}
	var sb strings.Builder
	sb.WriteString("🎯 *Daftar Rencana / To-Do Aktif:*\n\n")
	for i, t := range todos {
		sb.WriteString(fmt.Sprintf("%d. #%d — %s\n", i+1, t.ID, t.Title))
	}
	sb.WriteString("\nℹ️ Ketik `/todo done <id>` untuk menandai selesai.")
	return sb.String()
}

func RenderTodoDone(todo *domain.Todo) string {
	return fmt.Sprintf("✅ *To-Do #%d Selesai!*\n• %s\n\n📝 _Otomatis masuk ke catatan aktivitas hari ini._", todo.ID, todo.Title)
}

func RenderBlockerRecorded(blocker *domain.Blocker) string {
	return fmt.Sprintf("🚧 *Kendala Tercatat!*\n• ID: #%d\n• Blocker: %s\n\nℹ️ Ketik `/blocker done %d` jika kendala sudah beres.", blocker.ID, blocker.Description, blocker.ID)
}

func RenderBlockerList(blockers []*domain.Blocker) string {
	if len(blockers) == 0 {
		return "🎉 Tidak ada kendala aktif saat ini. Pekerjaan aman lancar!"
	}
	var sb strings.Builder
	sb.WriteString("🚧 *Daftar Kendala / Blockers Aktif:*\n\n")
	for i, b := range blockers {
		sb.WriteString(fmt.Sprintf("%d. #%d — %s\n", i+1, b.ID, b.Description))
	}
	sb.WriteString("\nℹ️ Ketik `/blocker done <id>` jika sudah terselesaikan.")
	return sb.String()
}

func RenderBlockerDone(blocker *domain.Blocker) string {
	return fmt.Sprintf("🎉 *Kendala #%d Terselesaikan!*\n• %s\n\nBagus! Jalur kodingan kembali lancar.", blocker.ID, blocker.Description)
}

func RenderRepoList(repos []*domain.MonitoredRepo) string {
	if len(repos) == 0 {
		return "📂 Belum ada repositori yang dipantau.\nTambah via `/repo add <owner/repo>` (misal: `/repo add AryaKhaslana/PocketPulse`)."
	}
	var sb strings.Builder
	sb.WriteString("📂 *Daftar Repositori yang Dipantau:*\n\n")
	for i, r := range repos {
		lastSync := "Belum pernah"
		if r.LastSyncedAt != nil {
			lastSync = r.LastSyncedAt.Format("02 Jan 15:04")
		}
		sb.WriteString(fmt.Sprintf("%d. `%s` (branch: %s) — Sync terakhir: %s\n", i+1, r.RepoName, r.Branch, lastSync))
	}
	sb.WriteString("\nℹ️ Ketik `/sync` untuk menarik commit terbaru sekarang.")
	return sb.String()
}

func RenderSearchResults(query string, logs []*domain.WorkLog) string {
	if len(logs) == 0 {
		return fmt.Sprintf("🔍 Tidak ditemukan catatan untuk kata kunci: `%s`", query)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 *Hasil Pencarian: `%s` (%d ditemukan)*\n\n", query, len(logs)))
	for i, l := range logs {
		repoPrefix := ""
		if l.RepoName != "" {
			repoPrefix = fmt.Sprintf("[%s] ", l.RepoName)
		}
		sb.WriteString(fmt.Sprintf("%d. `%s` %s %s%s\n", i+1, l.LogDate.Format("02 Jan"), l.Category.Icon(), repoPrefix, l.Title))
	}
	return sb.String()
}

func RenderContainerList(containers []docker.ContainerInfo) (string, InlineKeyboardMarkup) {
	if len(containers) == 0 {
		return "🐳 Tidak ada container Docker yang terdeteksi di server.", InlineKeyboardMarkup{}
	}

	var sb strings.Builder
	sb.WriteString("🐳 *Status Container Docker:*\n\n")

	var keyboard [][]InlineKeyboardButton
	for i, c := range containers {
		name := c.CleanName()
		statusIcon := "🟢"
		if !c.IsRunning() {
			statusIcon = "🔴"
		}

		sb.WriteString(fmt.Sprintf("%d. %s `%s`\n   • Image: `%s`\n   • Status: %s\n", i+1, statusIcon, name, c.Image, c.Status))

		actionRow := []InlineKeyboardButton{
			{Text: fmt.Sprintf("🔄 Restart %s", name), CallbackData: "docker:restart:" + name},
			{Text: fmt.Sprintf("📜 Log %s", name), CallbackData: "docker:logs:" + name},
		}
		keyboard = append(keyboard, actionRow)
	}

	sb.WriteString("\nℹ️ Klik tombol di atas atau ketik `/docker restart <nama>` / `/docker logs <nama>`.")
	return sb.String(), InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func RenderIdeaRecorded(idea *domain.Idea) string {
	return fmt.Sprintf("💡 *Ide Tersimpan di Idea Vault!*\n• ID: #%d\n• Ide: %s\n\nℹ️ Aman tersimpan! Ketik `/ide` untuk melihat daftar atau memindahkannya ke To-Do.", idea.ID, idea.Title)
}

func RenderIdeaList(ideas []*domain.Idea) (string, InlineKeyboardMarkup) {
	if len(ideas) == 0 {
		return "💡 *Idea Vault Kosong.*\nBelum ada ide yang disimpan. Ketik `ide: <gagasan kamu>` untuk mencatat ide liar tanpa mengotori to-do list harianmu!", InlineKeyboardMarkup{}
	}

	var sb strings.Builder
	sb.WriteString("💡 *Daftar Ide di Idea Vault / Backlog:*\n\n")

	var keyboard [][]InlineKeyboardButton
	for i, item := range ideas {
		sb.WriteString(fmt.Sprintf("%d. #%d — %s\n", i+1, item.ID, item.Title))

		row := []InlineKeyboardButton{
			{Text: fmt.Sprintf("🎯 Eksekusi #%d ke To-Do", item.ID), CallbackData: fmt.Sprintf("idea:promote:%d", item.ID)},
			{Text: fmt.Sprintf("🗑️ Hapus #%d", item.ID), CallbackData: fmt.Sprintf("idea:del:%d", item.ID)},
		}
		keyboard = append(keyboard, row)
	}

	sb.WriteString("\nℹ️ Klik tombol di bawah atau ketik `/ide todo <id>` untuk menjadikan to-do.")
	return sb.String(), InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func RenderIdeaPromoted(idea *domain.Idea, todo *domain.Todo) string {
	return fmt.Sprintf("🚀 *Ide #%d Dipromosikan ke To-Do #%d!*\n• %s\n\n🎯 Siap dieksekusi hari ini! Ketik `/todo` untuk melihat daftar kerjaan aktif.", idea.ID, todo.ID, todo.Title)
}
