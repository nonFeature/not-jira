package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"not-jira/internal/ai"
	"not-jira/internal/bot"
	"not-jira/internal/config"
	"not-jira/internal/storage/sqlite"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML configuration file")
	flag.Parse()

	log.Println("[Main] Starting not-jira...")

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("[Main FATAL] Configuration error: %v", err)
	}
	log.Printf("[Main] Loaded configuration from: %s", *configPath)

	st, err := sqlite.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("[Main FATAL] Database initialization error: %v", err)
	}
	defer st.Close()
	log.Printf("[Main] Database opened successfully: %s", cfg.Database.Path)

	var summarizer *ai.Summarizer
	if cfg.AI.Enabled {
		sm, err := ai.New(cfg.AI, cfg.Telegram.ProxyURL)
		if err != nil {
			log.Printf("[Main WARN] Failed to initialize AI summarizer: %v", err)
		} else {
			summarizer = sm
			log.Printf("[Main] AI summarizer enabled (model: %s, base: %s)", cfg.AI.Model, cfg.AI.BaseURL)
		}
	} else {
		log.Println("[Main] AI summarizer disabled. Manual interactive form will be used in DM.")
	}

	botService, err := bot.New(cfg, st, summarizer)
	if err != nil {
		log.Fatalf("[Main FATAL] Bot initialization error: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := botService.Start(ctx); err != nil {
		log.Fatalf("[Main FATAL] Bot runtime error: %v", err)
	}

	log.Println("[Main] not-jira shutdown cleanly.")
}
