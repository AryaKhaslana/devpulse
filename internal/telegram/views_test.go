package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"devpulse/internal/domain"
)

func TestRenderStart(t *testing.T) {
	out := RenderStart("AryaKhaslana")
	if !strings.Contains(out, "AryaKhaslana") {
		t.Errorf("Expected greeting for AryaKhaslana, got: %s", out)
	}
	if !strings.Contains(out, "/standup") {
		t.Errorf("Expected command /standup in help text")
	}
}

func TestRenderLogRecorded(t *testing.T) {
	l := &domain.WorkLog{
		Category: domain.CategoryCode,
		RepoName: "AryaKhaslana/PocketPulse",
		Title:    "feat: add scheduler",
		LogDate:  time.Now(),
	}
	out := RenderLogRecorded(l)
	if !strings.Contains(out, "Aktivitas Tercatat!") {
		t.Errorf("Expected header in output: %s", out)
	}
	if !strings.Contains(out, "PocketPulse") {
		t.Errorf("Expected repo name in output: %s", out)
	}
}

func TestRenderTodoWorkflow(t *testing.T) {
	todo := &domain.Todo{
		ID:        1,
		AccountID: uuid.New(),
		Title:     "Bikin unit test",
	}

	outRec := RenderTodoRecorded(todo)
	if !strings.Contains(outRec, "#1") || !strings.Contains(outRec, "Bikin unit test") {
		t.Errorf("Unexpected render todo recorded: %s", outRec)
	}

	outList := RenderTodoList([]*domain.Todo{todo})
	if !strings.Contains(outList, "#1 — Bikin unit test") {
		t.Errorf("Unexpected render todo list: %s", outList)
	}

	outDone := RenderTodoDone(todo)
	if !strings.Contains(outDone, "To-Do #1 Selesai!") {
		t.Errorf("Unexpected render todo done: %s", outDone)
	}
}
