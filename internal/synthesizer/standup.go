package synthesizer

import (
	"fmt"
	"strings"
	"time"

	"devpulse/internal/domain"
)

var indonesianDays = [...]string{
	"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu",
}

var indonesianMonths = [...]string{
	"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agt", "Sep", "Okt", "Nov", "Des",
}

func FormatIndonesianDate(t time.Time) string {
	dayName := indonesianDays[int(t.Weekday())]
	monthName := indonesianMonths[int(t.Month())]
	return fmt.Sprintf("%s, %02d %s %d", dayName, t.Day(), monthName, t.Year())
}

func GenerateDailyStandup(
	username string,
	date time.Time,
	yesterdayLogs []*domain.WorkLog,
	activeTodos []*domain.Todo,
	activeBlockers []*domain.Blocker,
) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("☕ *Daily Standup — %s*\n", FormatIndonesianDate(date)))
	if username != "" {
		sb.WriteString(fmt.Sprintf("👤 Engineer: @%s\n\n", username))
	} else {
		sb.WriteString("\n")
	}

	// 1. Yesterday
	sb.WriteString("⏮️ *Kemarin (Yesterday):*\n")
	if len(yesterdayLogs) == 0 {
		sb.WriteString("• Melakukan riset teknis dan koordinasi tugas backlog.\n")
	} else {
		for _, l := range yesterdayLogs {
			repoPrefix := ""
			if l.RepoName != "" {
				repoPrefix = fmt.Sprintf("[%s] ", l.RepoName)
			}
			sb.WriteString(fmt.Sprintf("• %s%s %s\n", l.Category.Icon(), repoPrefix, l.Title))
		}
	}
	sb.WriteString("\n")

	// 2. Today
	sb.WriteString("🎯 *Hari Ini (Today):*\n")
	if len(activeTodos) == 0 {
		sb.WriteString("• Melanjutkan implementasi tugas backlog dan review PR.\n")
	} else {
		for _, todo := range activeTodos {
			sb.WriteString(fmt.Sprintf("• %s\n", todo.Title))
		}
	}
	sb.WriteString("\n")

	// 3. Blockers
	sb.WriteString("🚧 *Kendala (Blockers):*\n")
	if len(activeBlockers) == 0 {
		sb.WriteString("• Nihil (Aman lancar) 🎉\n")
	} else {
		for _, b := range activeBlockers {
			sb.WriteString(fmt.Sprintf("• ⚠️ %s\n", b.Description))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}
