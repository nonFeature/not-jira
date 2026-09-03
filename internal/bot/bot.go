package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"not-jira/internal/ai"
	"not-jira/internal/config"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
	"golang.org/x/net/proxy"
)

type BotService struct {
	bot        *telego.Bot
	cfg        *config.Config
	storage    storage.Storage
	summarizer *ai.Summarizer
	fsm        *FSM

	addHandler  *AddHandler
	viewHandler *ViewHandler
	editHandler *EditHandler
	userHandler *UserHandler
	notifier    *Notifier
}

func New(cfg *config.Config, st storage.Storage, sm *ai.Summarizer) (*BotService, error) {
	httpClient := &http.Client{
		Timeout: 35 * time.Second,
	}

	if cfg.Telegram.ProxyURL != "" {
		pURL, err := url.Parse(cfg.Telegram.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}

		if pURL.Scheme == "socks5" {
			dialer, err := proxy.FromURL(pURL, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("failed to setup socks5 dialer: %w", err)
			}
			httpClient.Transport = &http.Transport{
				DialContext: dialer.(proxy.ContextDialer).DialContext,
			}
			log.Printf("[Bot] Using SOCKS5 proxy: %s", cfg.Telegram.ProxyURL)
		} else if pURL.Scheme == "http" || pURL.Scheme == "https" {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(pURL),
			}
			log.Printf("[Bot] Using HTTP proxy: %s", cfg.Telegram.ProxyURL)
		}
	}

	bot, err := telego.NewBot(cfg.Telegram.Token, telego.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telego bot: %w", err)
	}

	botUser, err := bot.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to call GetMe on telegram bot (check token / proxy): %w", err)
	}
	log.Printf("[Bot] Authenticated as @%s (ID: %d)", botUser.Username, botUser.ID)

	fsm := NewFSM()
	notifier := NewNotifier(bot, st)

	s := &BotService{
		bot:         bot,
		cfg:         cfg,
		storage:     st,
		summarizer:  sm,
		fsm:         fsm,
		addHandler:  NewAddHandler(bot, botUser.Username, cfg, st, sm, fsm),
		viewHandler: NewViewHandler(bot, botUser.Username, cfg, st),
		editHandler: NewEditHandler(bot, cfg, st, fsm, notifier),
		userHandler: NewUserHandler(bot, botUser.Username, cfg, st),
		notifier:    notifier,
	}

	return s, nil
}

func (s *BotService) Start(ctx context.Context) error {
	updates, err := s.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return fmt.Errorf("failed to start long polling: %w", err)
	}
	defer s.bot.StopLongPolling()

	log.Println("[Bot] Long polling started successfully. Ready to process updates asynchronously.")

	for {
		select {
		case <-ctx.Done():
			log.Println("[Bot] Stopping update processing...")
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			// Process each update asynchronously in a separate goroutine
			go s.dispatchUpdate(ctx, update)
		}
	}
}

func (s *BotService) dispatchUpdate(ctx context.Context, update telego.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Bot ERROR] Panic during update processing: %v", r)
		}
	}()

	// 1. Process Callback Queries (Buttons)
	if update.CallbackQuery != nil {
		query := update.CallbackQuery
		data := query.Data

		if data == "toggle_notify_dm" {
			s.userHandler.HandleToggleNotify(ctx, query)
			return
		}

		if strings.HasPrefix(data, "list:") || strings.HasPrefix(data, "view:") || data == "noop" {
			s.viewHandler.HandleCallback(ctx, query)
			return
		}

		s.editHandler.HandleCallback(ctx, query)
		return
	}

	// 2. Process Messages
	if update.Message != nil {
		msg := update.Message

		// First check if user is in an active FSM session (interactive form)
		if s.editHandler.HandleFSMMessage(ctx, msg) {
			return
		}

		text := strings.TrimSpace(msg.Text)
		cmd := strings.Split(text, " ")[0]
		cmd = strings.ToLower(cmd)
		// Strip bot username if mentioned (e.g. /list@not_jira_bot -> /list)
		if atIdx := strings.Index(cmd, "@"); atIdx != -1 {
			cmd = cmd[:atIdx]
		}

		switch cmd {
		case "/start":
			s.userHandler.HandleStart(ctx, msg)
		case "/help":
			s.userHandler.HandleStart(ctx, msg)
		case "/settings":
			s.userHandler.HandleSettings(ctx, msg)
		case "/add":
			s.addHandler.Handle(ctx, msg)
		case "/list":
			s.viewHandler.HandleList(ctx, msg)
		case "/view":
			s.viewHandler.HandleView(ctx, msg)
		}
	}
}
