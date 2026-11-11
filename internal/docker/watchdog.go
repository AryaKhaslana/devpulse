package docker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"devpulse/internal/domain"
)

type Notifier interface {
	SendMessage(chatID int64, text string) error
	SendMessageWithKeyboard(chatID int64, text string, markup interface{}) error
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type Watchdog struct {
	client      Client
	accountRepo domain.AccountRepository
	notifier    Notifier
	mu          sync.Mutex
	lastState   map[string]string
	firstRun    bool
}

func NewWatchdog(client Client, accountRepo domain.AccountRepository, notifier Notifier) *Watchdog {
	return &Watchdog{
		client:      client,
		accountRepo: accountRepo,
		notifier:    notifier,
		lastState:   make(map[string]string),
		firstRun:    true,
	}
}

func (w *Watchdog) Start(ctx context.Context, interval time.Duration) {
	if interval < 5*time.Second {
		interval = 15 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Docker Watchdog] Stopped.")
				return
			case <-ticker.C:
				w.Check(ctx)
			}
		}
	}()
	log.Printf("[Docker Watchdog] Active. Polling container health every %v", interval)
}

func (w *Watchdog) Check(ctx context.Context) {
	containers, err := w.client.ListContainers(ctx, true)
	if err != nil {
		// Docker daemon might be temporarily unavailable
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.firstRun {
		for _, c := range containers {
			w.lastState[c.CleanName()] = c.State
		}
		w.firstRun = false
		return
	}

	accounts, err := w.accountRepo.GetAll(ctx)
	if err != nil || len(accounts) == 0 {
		return
	}

	for _, c := range containers {
		name := c.CleanName()
		prevState, exists := w.lastState[name]

		// Detected crash: was running, now exited or dead
		if exists && prevState == "running" && (c.State == "exited" || c.State == "dead") {
			log.Printf("[Docker Watchdog] CRASH DETECTED on container '%s' (status: %s)", name, c.Status)

			alertMsg := fmt.Sprintf(
				"🚨 *PERINGATAN DOCKER WATCHDOG!*\n\n"+
					"Container `%s` mendadak *MATI / CRASH*!\n"+
					"• Status: `%s`\n"+
					"• State: `%s`\n\n"+
					"Gunakan tombol di bawah untuk tindakan cepat:",
				name, c.Status, c.State,
			)

			markup := InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{
						{Text: fmt.Sprintf("🔄 Restart %s", name), CallbackData: "docker:restart:" + name},
						{Text: fmt.Sprintf("📜 Cek Log %s", name), CallbackData: "docker:logs:" + name},
					},
				},
			}

			for _, acc := range accounts {
				if w.notifier != nil {
					_ = w.notifier.SendMessageWithKeyboard(acc.TelegramChatID, alertMsg, markup)
				}
			}
		}

		w.lastState[name] = c.State
	}
}
