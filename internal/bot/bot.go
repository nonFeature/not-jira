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

	botUser, err := bot.GetMe(context.Background())
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
	pollCtx, pollCancel := context.WithCancel(ctx)
	defer pollCancel()

	updates, err := s.bot.UpdatesViaLongPolling(pollCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to start long polling: %w", err)
	}

	log.Println("[Bot] Long polling started successfully. Ready to process updates asynchronously.")

	// Start auto-archiver background worker (silent, checks every hour)
	go s.startAutoArchiver(ctx)

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

func (s *BotService) startAutoArchiver(ctx context.Context) {
	// Initial check on startup
	s.runAutoArchive(ctx)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runAutoArchive(ctx)
		}
	}
}

func (s *BotService) runAutoArchive(ctx context.Context) {
	count, err := s.storage.ArchiveInactiveClosedTasks(ctx, 7*24*time.Hour)
	if err != nil {
		log.Printf("[Archiver ERROR] Auto-archive failed: %v", err)
		return
	}
	if count > 0 {
		log.Printf("[Archiver] Auto-archived %d inactive closed task(s) (> 7 days).", count)
	}
}

func (s *BotService) dispatchUpdate(ctx context.Context, update telego.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Bot ERROR] Panic during update processing: %v", r)
		}
	}()

	// Cache user profile for lookups and delegation
	if update.Message != nil && update.Message.From != nil {
		from := update.Message.From
		_ = s.storage.UpsertUser(ctx, from.ID, from.Username, from.FirstName)
	} else if update.CallbackQuery != nil {
		from := &update.CallbackQuery.From
		_ = s.storage.UpsertUser(ctx, from.ID, from.Username, from.FirstName)
	}

	// 1. Process Callback Queries (Buttons)
	if update.CallbackQuery != nil {
		query := update.CallbackQuery
		data := query.Data

		if strings.HasPrefix(data, "settings:") || data == "toggle_notify_dm" {
			s.userHandler.HandleToggleNotify(ctx, query)
			return
		}

		if strings.HasPrefix(data, "list:") || strings.HasPrefix(data, "list_tags:") || strings.HasPrefix(data, "view:") || strings.HasPrefix(data, "my:") || data == "noop" {
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
		case "/settings":
			s.userHandler.HandleSettings(ctx, msg)
		case "/add":
			s.addHandler.Handle(ctx, msg)
		case "/list":
			s.viewHandler.HandleList(ctx, msg)
		case "/my":
			s.viewHandler.HandleMyTasks(ctx, msg)
		case "/view":
			s.viewHandler.HandleView(ctx, msg)
		case "/cancel":
			s.editHandler.HandleCancel(ctx, msg)
		}
	}
}
