package config

import (
	"os"
	"time"
)

type Config struct {
	BotToken           string
	DatabaseURL        string
	Timezone           string
	Location           *time.Location
	BotMode            string // "polling" or "webhook"
	WebhookURL         string
	WebhookPort        string
	WebhookPath        string
	GitHubDefaultToken string
}

func Load() (*Config, error) {
	tz := getEnv("TIMEZONE", "Asia/Jakarta")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	cfg := &Config{
		BotToken:           getEnv("BOT_TOKEN", ""),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://devpulse:devpulse_secret@localhost:5433/devpulse?sslmode=disable"),
		Timezone:           tz,
		Location:           loc,
		BotMode:            getEnv("BOT_MODE", "polling"),
		WebhookURL:         getEnv("WEBHOOK_URL", ""),
		WebhookPort:        getEnv("WEBHOOK_PORT", "8090"),
		WebhookPath:        getEnv("WEBHOOK_PATH", "/webhook"),
		GitHubDefaultToken: getEnv("GITHUB_DEFAULT_TOKEN", ""),
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
