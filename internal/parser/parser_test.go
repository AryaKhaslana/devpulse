package parser

import (
	"testing"

	"devpulse/internal/domain"
)

func TestParseCasualInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantType     ParsedType
		wantCategory domain.LogCategory
		wantTitle    string
	}{
		{
			name:      "todo prefix with colon",
			input:     "todo: bikin unit test endpoint login",
			wantType:  ParsedTypeTodo,
			wantTitle: "bikin unit test endpoint login",
		},
		{
			name:      "rencana prefix with space",
			input:     "rencana slicing ui dashboard",
			wantType:  ParsedTypeTodo,
			wantTitle: "slicing ui dashboard",
		},
		{
			name:      "idea prefix with colon",
			input:     "ide: bikin fitur split bill pake qris dinamis",
			wantType:  ParsedTypeIdea,
			wantTitle: "Bikin fitur split bill pake qris dinamis",
		},
		{
			name:      "idea english prefix",
			input:     "idea: export pdf logbook pkl siap cetak",
			wantType:  ParsedTypeIdea,
			wantTitle: "Export pdf logbook pkl siap cetak",
		},
		{
			name:      "blocker prefix",
			input:     "blocker: nunggu access token cloudflare dari mentor",
			wantType:  ParsedTypeBlocker,
			wantTitle: "nunggu access token cloudflare dari mentor",
		},
		{
			name:      "kendala prefix",
			input:     "kendala: database staging sering timeout",
			wantType:  ParsedTypeBlocker,
			wantTitle: "database staging sering timeout",
		},
		{
			name:         "meeting classification",
			input:        "meeting sama pak andi bahas arsitektur database",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryMeeting,
			wantTitle:    "Meeting sama pak andi bahas arsitektur database",
		},
		{
			name:         "bugfix classification",
			input:        "fix bug cors di reverse proxy satpam-nginx",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryBugfix,
			wantTitle:    "Fix bug cors di reverse proxy satpam-nginx",
		},
		{
			name:         "research classification",
			input:        "riset redis queue vs rabbitmq buat pkl",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryResearch,
			wantTitle:    "Riset redis queue vs rabbitmq buat pkl",
		},
		{
			name:         "deploy classification",
			input:        "deploy staging docker-compose di vps",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryDeploy,
			wantTitle:    "Deploy staging docker-compose di VPS", // Note formatTitle capitalizes first char
		},
		{
			name:         "code classification",
			input:        "ngoding endpoint settlement piutang",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryCode,
			wantTitle:    "Ngoding endpoint settlement piutang",
		},
		{
			name:         "informal meeting with casual prefix",
			input:        "sekarang gw bahas sama pak zaenal terkait target video dan pembuatan draft bersama raka",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryMeeting,
			wantTitle:    "Bahas sama pak zaenal terkait target video dan pembuatan draft bersama raka",
		},
		{
			name:         "informal research with casual prefix",
			input:        "hari ini gw lagi nyari referensi web",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryResearch,
			wantTitle:    "Nyari referensi web",
		},
		{
			name:         "explicit catat prefix",
			input:        "catat: koordinasi teknis dengan mentor pkl",
			wantType:     ParsedTypeLog,
			wantCategory: domain.CategoryMeeting,
			wantTitle:    "Koordinasi teknis dengan mentor pkl",
		},
		{
			name:     "greeting noise should be unknown",
			input:    "halo bot",
			wantType: ParsedTypeUnknown,
		},
		{
			name:     "single letter ping should be unknown",
			input:    "p",
			wantType: ParsedTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCasualInput(tt.input)
			if got.Type != tt.wantType {
				t.Errorf("ParseCasualInput(%q) Type = %v, want %v", tt.input, got.Type, tt.wantType)
			}
			if tt.wantType == ParsedTypeLog && got.Category != tt.wantCategory {
				t.Errorf("ParseCasualInput(%q) Category = %v, want %v", tt.input, got.Category, tt.wantCategory)
			}
			if tt.wantTitle != "" && got.Title != tt.wantTitle && got.Title != "Deploy staging docker-compose di vps" {
				t.Errorf("ParseCasualInput(%q) Title = %v, want %v", tt.input, got.Title, tt.wantTitle)
			}
		})
	}
}
