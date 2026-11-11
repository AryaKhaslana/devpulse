package synthesizer

import (
	"fmt"
	"strings"
	"time"

	"devpulse/internal/domain"
)

func GenerateWeeklyLogbook(
	username string,
	mondayDate time.Time,
	fridayDate time.Time,
	logs []*domain.WorkLog,
	loc *time.Location,
) string {
	var sb strings.Builder

	sb.WriteString("📋 *LOGBOOK MINGGUAN PKL / KERJA*\n")
	sb.WriteString(fmt.Sprintf("📅 Periode: %02d %s %d — %02d %s %d\n",
		mondayDate.Day(), indonesianMonths[int(mondayDate.Month())], mondayDate.Year(),
		fridayDate.Day(), indonesianMonths[int(fridayDate.Month())], fridayDate.Year(),
	))
	if username != "" {
		sb.WriteString(fmt.Sprintf("👤 Praktikan/Engineer: @%s\n\n", username))
	} else {
		sb.WriteString("\n")
	}

	// Group logs by YYYY-MM-DD
	grouped := make(map[string][]*domain.WorkLog)
	catCounts := make(map[domain.LogCategory]int)

	for _, l := range logs {
		key := l.LogDate.In(loc).Format("2006-01-02")
		grouped[key] = append(grouped[key], l)
		catCounts[l.Category]++
	}

	// Iterate 5 working days (Monday to Friday)
	cur := mondayDate
	hasAny := false
	for i := 0; i < 5; i++ {
		dateStr := cur.Format("2006-01-02")
		dayHeader := fmt.Sprintf("%d. %s", i+1, FormatIndonesianDate(cur))
		sb.WriteString(fmt.Sprintf("*%s*\n", dayHeader))

		dayLogs := grouped[dateStr]
		if len(dayLogs) == 0 {
			sb.WriteString("   _Tidak ada aktivitas koding/catatan terekam._\n\n")
		} else {
			hasAny = true
			for _, l := range dayLogs {
				repoPrefix := ""
				if l.RepoName != "" {
					repoPrefix = fmt.Sprintf("[%s] ", l.RepoName)
				}
				sb.WriteString(fmt.Sprintf("   • %s %s%s\n", l.Category.Icon(), repoPrefix, l.Title))
			}
			sb.WriteString("\n")
		}

		cur = cur.AddDate(0, 0, 1)
	}

	sb.WriteString("─────────────────────────────\n")
	if !hasAny {
		sb.WriteString("ℹ️ Belum ada aktivitas yang tercatat pada rentang minggu ini.\n")
	} else {
		sb.WriteString(fmt.Sprintf("📊 *Total Capaian: %d Aktivitas*\n", len(logs)))
		if catCounts[domain.CategoryCode] > 0 {
			sb.WriteString(fmt.Sprintf("• 💻 Koding/Git: %d\n", catCounts[domain.CategoryCode]))
		}
		if catCounts[domain.CategoryMeeting] > 0 {
			sb.WriteString(fmt.Sprintf("• 👥 Meeting/Koordinasi: %d\n", catCounts[domain.CategoryMeeting]))
		}
		if catCounts[domain.CategoryBugfix] > 0 {
			sb.WriteString(fmt.Sprintf("• 🐛 Bugfix: %d\n", catCounts[domain.CategoryBugfix]))
		}
		if catCounts[domain.CategoryResearch] > 0 {
			sb.WriteString(fmt.Sprintf("• 🔍 Riset/Belajar: %d\n", catCounts[domain.CategoryResearch]))
		}
		if catCounts[domain.CategoryDeploy] > 0 {
			sb.WriteString(fmt.Sprintf("• 🚀 Deploy/DevOps: %d\n", catCounts[domain.CategoryDeploy]))
		}
		if catCounts[domain.CategoryTask] > 0 {
			sb.WriteString(fmt.Sprintf("• 📝 Tugas/Lainnya: %d\n", catCounts[domain.CategoryTask]))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func GenerateDailySummary(date time.Time, logs []*domain.WorkLog) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📅 *Timeline Aktivitas Hari Ini — %s*\n\n", FormatIndonesianDate(date)))

	if len(logs) == 0 {
		sb.WriteString("Belum ada aktivitas yang tercatat hari ini.\n")
		sb.WriteString("Ketik chat santai (misal: _'meeting sama mentor'_) atau commit kodingan di GitHub.")
		return sb.String()
	}

	for i, l := range logs {
		repoPrefix := ""
		if l.RepoName != "" {
			repoPrefix = fmt.Sprintf("[%s] ", l.RepoName)
		}
		timeStr := l.CreatedAt.Format("15:04")
		sb.WriteString(fmt.Sprintf("%d. `%s` %s %s%s\n", i+1, timeStr, l.Category.Icon(), repoPrefix, l.Title))
	}

	sb.WriteString(fmt.Sprintf("\n📊 Total: %d aktivitas terekam.", len(logs)))
	return sb.String()
}
