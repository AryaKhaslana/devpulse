package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"devpulse/internal/config"
)

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	From      *User  `json:"from"`
	Text      string `json:"text"`
	Date      int64  `json:"date"`
}

type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type SendMessageRequest struct {
	ChatID      int64       `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

type Bot struct {
	token       string
	pollClient  *http.Client
	sendClient  *http.Client
	handler     *Handler
	mode        string
	webhookURL  string
	webhookPort string
	webhookPath string
}

func NewBot(cfg *config.Config, handler *Handler) *Bot {
	pollTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 35 * time.Second,
	}

	sendTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
	}

	return &Bot{
		token: cfg.BotToken,
		pollClient: &http.Client{
			Transport: pollTransport,
			Timeout:   40 * time.Second,
		},
		sendClient: &http.Client{
			Transport: sendTransport,
			Timeout:   15 * time.Second,
		},
		handler:     handler,
		mode:        cfg.BotMode,
		webhookURL:  cfg.WebhookURL,
		webhookPort: cfg.WebhookPort,
		webhookPath: cfg.WebhookPath,
	}
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	if b.token == "" {
		log.Printf("[Bot MOCK Send] Chat %d: %s", chatID, text)
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	payload := SendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal sendMessage request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := b.sendClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}

		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Retry without markdown if markdown parsing failed
			if bytes.Contains(respBytes, []byte("can't parse entities")) {
				payload.ParseMode = ""
				plainBody, _ := json.Marshal(payload)
				reqPlain, _ := http.NewRequest("POST", url, bytes.NewBuffer(plainBody))
				reqPlain.Header.Set("Content-Type", "application/json")
				respPlain, errPlain := b.sendClient.Do(reqPlain)
				if errPlain == nil {
					respPlain.Body.Close()
					if respPlain.StatusCode == http.StatusOK {
						return nil
					}
				}
			}

			lastErr = fmt.Errorf("telegram API status %d: %s", resp.StatusCode, string(respBytes))
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}

		log.Printf("[Bot] Successfully sent reply to chat %d", chatID)
		return nil
	}

	return fmt.Errorf("failed to send message after 3 attempts: %w", lastErr)
}

func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, markup interface{}) error {
	if b.token == "" {
		log.Printf("[Bot MOCK Send Keyboard] Chat %d: %s", chatID, text)
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	payload := SendMessageRequest{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "Markdown",
		ReplyMarkup: markup,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal sendMessage request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := b.sendClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}

		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if bytes.Contains(respBytes, []byte("can't parse entities")) {
				payload.ParseMode = ""
				plainBody, _ := json.Marshal(payload)
				reqPlain, _ := http.NewRequest("POST", url, bytes.NewBuffer(plainBody))
				reqPlain.Header.Set("Content-Type", "application/json")
				respPlain, errPlain := b.sendClient.Do(reqPlain)
				if errPlain == nil {
					respPlain.Body.Close()
					if respPlain.StatusCode == http.StatusOK {
						return nil
					}
				}
			}

			lastErr = fmt.Errorf("telegram API status %d: %s", resp.StatusCode, string(respBytes))
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
			continue
		}

		log.Printf("[Bot] Successfully sent keyboard reply to chat %d", chatID)
		return nil
	}

	return fmt.Errorf("failed to send message with keyboard after 3 attempts: %w", lastErr)
}

type AnswerCallbackQueryRequest struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
}

func (b *Bot) AnswerCallbackQuery(callbackQueryID string, text string, showAlert bool) error {
	if b.token == "" {
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", b.token)
	payload := AnswerCallbackQueryRequest{
		CallbackQueryID: callbackQueryID,
		Text:            text,
		ShowAlert:       showAlert,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.sendClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (b *Bot) Start(ctx context.Context) error {
	if b.token == "" {
		log.Printf("[Bot] Warning: BOT_TOKEN is not set. Bot running in offline mode.")
		<-ctx.Done()
		return nil
	}

	if b.mode == "webhook" && b.webhookURL != "" {
		return b.startWebhook(ctx)
	}

	return b.startPolling(ctx)
}

func (b *Bot) startWebhook(ctx context.Context) error {
	log.Printf("[Bot] Starting in Webhook mode. Target: %s", b.webhookURL)

	mux := http.NewServeMux()
	mux.HandleFunc(b.webhookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var update Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if update.Message != nil {
			go func(m *Message) {
				if err := b.handler.Dispatch(context.Background(), m); err != nil {
					log.Printf("[Bot] Handler error: %v", err)
				}
			}(update.Message)
		} else if update.CallbackQuery != nil {
			go func(cb *CallbackQuery) {
				if err := b.handler.DispatchCallback(context.Background(), cb); err != nil {
					log.Printf("[Bot] Callback error: %v", err)
				}
			}(update.CallbackQuery)
		}

		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    ":" + b.webhookPort,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("[Bot] Webhook listening on port %s%s", b.webhookPort, b.webhookPath)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("webhook server error: %w", err)
	}

	return nil
}

func (b *Bot) startPolling(ctx context.Context) error {
	log.Printf("[Bot] Starting in Long Polling mode...")

	offset := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("[Bot] Polling loop exiting...")
			return nil
		default:
		}

		updates, err := b.getUpdates(offset, 10)
		if err != nil {
			log.Printf("[Bot] Polling notice: %v. Retrying in 2s...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}

			if u.Message != nil {
				msg := u.Message
				log.Printf("[Bot Polling] Received update from chat %d: %q", msg.Chat.ID, msg.Text)
				go func(m *Message) {
					if err := b.handler.Dispatch(context.Background(), m); err != nil {
						log.Printf("[Bot] Handler error for chat %d: %v", m.Chat.ID, err)
					}
				}(msg)
			} else if u.CallbackQuery != nil {
				cb := u.CallbackQuery
				log.Printf("[Bot Polling] Received callback query from user %d: %q", cb.From.ID, cb.Data)
				go func(c *CallbackQuery) {
					if err := b.handler.DispatchCallback(context.Background(), c); err != nil {
						log.Printf("[Bot] Callback error: %v", err)
					}
				}(cb)
			}
		}
	}
}

func (b *Bot) getUpdates(offset int, timeout int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d", b.token, offset, timeout)
	resp, err := b.pollClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK          bool     `json:"ok"`
		Result      []Update `json:"result"`
		Description string   `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	return apiResp.Result, nil
}
