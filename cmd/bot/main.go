package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devpulse/internal/config"
	"devpulse/internal/database"
	"devpulse/internal/docker"
	"devpulse/internal/github"
	"devpulse/internal/repository/postgres"
	"devpulse/internal/scheduler"
	"devpulse/internal/telegram"
)

func main() {
	log.Println("🚀 Initializing DevPulse Microservice...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Database & Run Migrations
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database connected & schema migrated successfully.")

	// Instantiate Repositories
	accountRepo := postgres.NewAccountRepo(db.Pool)
	repoRepo := postgres.NewMonitoredRepoRepo(db.Pool)
	workLogRepo := postgres.NewWorkLogRepo(db.Pool)
	todoRepo := postgres.NewTodoRepo(db.Pool)
	blockerRepo := postgres.NewBlockerRepo(db.Pool)
	ideaRepo := postgres.NewIdeaRepo(db.Pool)

	// Instantiate GitHub Client & Docker Client
	githubClient := github.NewClient()
	dockerClient := docker.NewClient("/var/run/docker.sock")

	// Instantiate Handler & Telegram Bot
	handler := telegram.NewHandler(accountRepo, repoRepo, workLogRepo, todoRepo, blockerRepo, ideaRepo, githubClient, dockerClient, nil, cfg.Location)
	bot := telegram.NewBot(cfg, handler)
	handler.SetSender(bot)

	// Initialize Docker Crash Watchdog (polls every 15s)
	watchdog := docker.NewWatchdog(dockerClient, accountRepo, bot)
	watchdog.Start(ctx, 15*time.Second)

	// Initialize Scheduler (Standup 08:30, Logbook Fri 17:00, Git Sync)
	sched := scheduler.NewScheduler(accountRepo, repoRepo, workLogRepo, todoRepo, blockerRepo, githubClient, bot, cfg.Location)
	sched.Start(ctx)

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received termination signal: %v. Initiating graceful shutdown...", sig)
		cancel()
		sched.Stop()
	}()

	log.Println("⚡ DevPulse is ready and listening for events.")
	if err := bot.Start(ctx); err != nil {
		log.Printf("Bot shutdown with error: %v", err)
	}
	log.Println("DevPulse stopped cleanly.")
}
