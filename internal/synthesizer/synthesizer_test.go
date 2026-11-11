package synthesizer

import (
	"strings"
	"testing"
	"time"

	"devpulse/internal/domain"
)

func TestGenerateDailyStandup(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)
	now := time.Date(2026, 11, 11, 8, 30, 0, 0, loc)

	yesterdayLogs := []*domain.WorkLog{
		{
			Category: domain.CategoryCode,
			RepoName: "AryaKhaslana/PocketPulse",
			Title:    "feat(dashboard): add live Mini App dashboard",
		},
		{
			Category: domain.CategoryMeeting,
			Title:    "Briefing arsitektur database dengan Pak Andi",
		},
	}

	activeTodos := []*domain.Todo{
		{Title: "Bikin unit test endpoint login"},
		{Title: "Slicing UI dashboard"},
	}

	activeBlockers := []*domain.Blocker{
		{Description: "Nunggu access token cloudflare dari mentor"},
	}

	res := GenerateDailyStandup("AryaKhaslana", now, yesterdayLogs, activeTodos, activeBlockers)

	if !strings.Contains(res, "Daily Standup — Rabu, 11 Nov 2026") {
		t.Errorf("Expected title with date, got:\n%s", res)
	}
	if !strings.Contains(res, "@AryaKhaslana") {
		t.Errorf("Expected username, got:\n%s", res)
	}
	if !strings.Contains(res, "feat(dashboard)") {
		t.Errorf("Expected commit title, got:\n%s", res)
	}
	if !strings.Contains(res, "Briefing arsitektur") {
		t.Errorf("Expected meeting title, got:\n%s", res)
	}
	if !strings.Contains(res, "Bikin unit test endpoint login") {
		t.Errorf("Expected todo, got:\n%s", res)
	}
	if !strings.Contains(res, "Nunggu access token cloudflare") {
		t.Errorf("Expected blocker, got:\n%s", res)
	}
}

func TestGenerateWeeklyLogbook(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)
	monday := time.Date(2026, 11, 9, 0, 0, 0, 0, loc)
	friday := time.Date(2026, 11, 13, 0, 0, 0, 0, loc)

	logs := []*domain.WorkLog{
		{
			Category: domain.CategoryCode,
			RepoName: "PocketPulse",
			Title:    "feat: runway analytics",
			LogDate:  monday,
		},
		{
			Category: domain.CategoryMeeting,
			Title:    "Standup mingguan dengan tim",
			LogDate:  monday,
		},
	}

	res := GenerateWeeklyLogbook("AryaKhaslana", monday, friday, logs, loc)

	if !strings.Contains(res, "LOGBOOK MINGGUAN PKL / KERJA") {
		t.Errorf("Expected header, got:\n%s", res)
	}
	if !strings.Contains(res, "09 Nov 2026 — 13 Nov 2026") {
		t.Errorf("Expected date range, got:\n%s", res)
	}
	if !strings.Contains(res, "runway analytics") {
		t.Errorf("Expected log entry, got:\n%s", res)
	}
	if !strings.Contains(res, "Total Capaian: 2 Aktivitas") {
		t.Errorf("Expected summary total, got:\n%s", res)
	}
}
