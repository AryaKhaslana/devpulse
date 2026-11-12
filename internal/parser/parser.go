package parser

import (
	"regexp"
	"strings"

	"devpulse/internal/domain"
)

type ParsedType string

const (
	ParsedTypeLog     ParsedType = "LOG"
	ParsedTypeTodo    ParsedType = "TODO"
	ParsedTypeIdea    ParsedType = "IDEA"
	ParsedTypeBlocker ParsedType = "BLOCKER"
	ParsedTypeUnknown ParsedType = "UNKNOWN"
)

type ParseResult struct {
	Type     ParsedType
	Category domain.LogCategory
	Title    string
	Details  string
}

var (
	todoRegex    = regexp.MustCompile(`(?i)^(?:todo|rencana|plan|target)[\s:]+(.+)`)
	ideaRegex    = regexp.MustCompile(`(?i)^(?:ide|idea|gagasan|inspo)[\s:]+(.+)`)
	blockerRegex = regexp.MustCompile(`(?i)^(?:blocker|kendala|hambatan|masalah)[\s:]+(.+)`)
	catatRegex   = regexp.MustCompile(`(?i)^(?:catat)[\s:]+(.+)`)

	casualPrefixRegex = regexp.MustCompile(`(?i)^(?:(?:tadi|hari\s+ini|sekarang|gw|lagi|habis|udah)\s+)+`)

	meetingWords  = []string{"meeting", "meet", "meetup", "briefing", "rapat", "diskusi", "ngobrol", "sync", "evaluasi", "koordinasi", "bahas", "ketemu", "bicara", "konsultasi", "interview", "pembahasan"}
	bugfixWords   = []string{"fix", "fixing", "bug", "error", "hotfix", "troubleshoot", "resolve", "benerin", "solved"}
	researchWords = []string{"riset", "research", "nyari", "cari", "referensi", "baca", "belajar", "pelajari", "eksplor", "explore", "benchmark", "analisa", "analisis", "investigasi", "browsing"}
	deployWords   = []string{"deploy", "deployment", "release", "publish", "vps", "docker", "ci/cd", "staging", "production"}
	codeWords     = []string{"ngoding", "koding", "coding", "bikin", "buat", "refactor", "fitur", "feature", "slicing", "implement", "implementasi", "endpoint", "api", "crud", "query", "migrasi", "desain", "design", "wireframe", "video", "editing", "draft"}

	greetings = []string{"halo", "hai", "hello", "hi", "p", "ping", "test", "tes", "selamat pagi", "selamat siang", "selamat malam", "assalamualaikum"}
)

func ParseCasualInput(input string) ParseResult {
	clean := strings.TrimSpace(input)
	if clean == "" {
		return ParseResult{Type: ParsedTypeUnknown}
	}

	// 1. Check Todo prefix
	if m := todoRegex.FindStringSubmatch(clean); len(m) > 1 {
		return ParseResult{
			Type:  ParsedTypeTodo,
			Title: strings.TrimSpace(m[1]),
		}
	}

	// 2. Check Idea prefix (Brain Dump / Idea Vault)
	if m := ideaRegex.FindStringSubmatch(clean); len(m) > 1 {
		return ParseResult{
			Type:  ParsedTypeIdea,
			Title: formatTitle(m[1]),
		}
	}

	// 3. Check Blocker prefix
	if m := blockerRegex.FindStringSubmatch(clean); len(m) > 1 {
		return ParseResult{
			Type:  ParsedTypeBlocker,
			Title: strings.TrimSpace(m[1]),
		}
	}

	// 3. Check Explicit Catat prefix
	if m := catatRegex.FindStringSubmatch(clean); len(m) > 1 {
		text := strings.TrimSpace(m[1])
		cat := classifyCategory(text)
		return ParseResult{
			Type:     ParsedTypeLog,
			Category: cat,
			Title:    formatTitle(text),
			Details:  text,
		}
	}

	// 4. Filter greetings / noise
	lower := strings.ToLower(clean)
	for _, g := range greetings {
		if lower == g || strings.HasPrefix(lower, g+" ") {
			return ParseResult{Type: ParsedTypeUnknown}
		}
	}

	// If less than 2 words and not recognized, treat as unknown
	words := strings.Fields(clean)
	if len(words) < 2 {
		return ParseResult{Type: ParsedTypeUnknown}
	}

	// 5. Casual Work Log classification
	cat := classifyCategory(clean)
	return ParseResult{
		Type:     ParsedTypeLog,
		Category: cat,
		Title:    formatTitle(clean),
		Details:  clean,
	}
}

func classifyCategory(text string) domain.LogCategory {
	lower := strings.ToLower(text)

	for _, w := range meetingWords {
		if strings.Contains(lower, w) {
			return domain.CategoryMeeting
		}
	}
	for _, w := range bugfixWords {
		if strings.Contains(lower, w) {
			return domain.CategoryBugfix
		}
	}
	for _, w := range researchWords {
		if strings.Contains(lower, w) {
			return domain.CategoryResearch
		}
	}
	for _, w := range deployWords {
		if strings.Contains(lower, w) {
			return domain.CategoryDeploy
		}
	}
	for _, w := range codeWords {
		if strings.Contains(lower, w) {
			return domain.CategoryCode
		}
	}

	return domain.CategoryTask
}

func formatTitle(text string) string {
	t := strings.TrimSpace(text)
	if len(t) == 0 {
		return ""
	}
	cleaned := strings.TrimSpace(casualPrefixRegex.ReplaceAllString(t, ""))
	if len(cleaned) >= 3 {
		t = cleaned
	}
	// Capitalize first character
	r := []rune(t)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}
