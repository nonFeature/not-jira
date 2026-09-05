package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Database DatabaseConfig `yaml:"database"`
	AI       AIConfig       `yaml:"ai"`
	Labels   []string       `yaml:"labels"`
}

type TelegramConfig struct {
	Token        string  `yaml:"token"`
	ProxyURL     string  `yaml:"proxy_url"`
	AdminIDs     []int64 `yaml:"admin_ids"`
	DevIDs       []int64 `yaml:"dev_ids"`
	BugsTopicID  int     `yaml:"bugs_topic_id"`
	IdeasTopicID int     `yaml:"ideas_topic_id"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type AIConfig struct {
	Enabled  bool   `yaml:"enabled"`
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	UseProxy bool   `yaml:"use_proxy"`
}

func (c *TelegramConfig) IsAdmin(userID int64) bool {
	for _, id := range c.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (c *TelegramConfig) IsDev(userID int64) bool {
	if c.IsAdmin(userID) {
		return true
	}
	for _, id := range c.DevIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Path: "bot.db",
		},
		AI: AIConfig{
			Enabled:  false,
			BaseURL:  "https://api.mistral.ai/v1",
			Model:    "mistral-small-latest",
			UseProxy: true,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) || os.Getenv("TELEGRAM_TOKEN") == "" {
			return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
		}
	}

	// Environment overrides
	if token := os.Getenv("TELEGRAM_TOKEN"); token != "" {
		cfg.Telegram.Token = token
	}
	if proxy := os.Getenv("TELEGRAM_PROXY_URL"); proxy != "" {
		cfg.Telegram.ProxyURL = proxy
	}
	if admins := os.Getenv("TELEGRAM_ADMIN_IDS"); admins != "" {
		cfg.Telegram.AdminIDs = nil
		for _, part := range strings.Split(admins, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.ParseInt(part, 10, 64); err == nil {
				cfg.Telegram.AdminIDs = append(cfg.Telegram.AdminIDs, id)
			}
		}
	}
	if devs := os.Getenv("TELEGRAM_DEV_IDS"); devs != "" {
		cfg.Telegram.DevIDs = nil
		for _, part := range strings.Split(devs, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.ParseInt(part, 10, 64); err == nil {
				cfg.Telegram.DevIDs = append(cfg.Telegram.DevIDs, id)
			}
		}
	}
	if bugsTopic := os.Getenv("TELEGRAM_BUGS_TOPIC_ID"); bugsTopic != "" {
		if id, err := strconv.Atoi(bugsTopic); err == nil {
			cfg.Telegram.BugsTopicID = id
		}
	}
	if ideasTopic := os.Getenv("TELEGRAM_IDEAS_TOPIC_ID"); ideasTopic != "" {
		if id, err := strconv.Atoi(ideasTopic); err == nil {
			cfg.Telegram.IdeasTopicID = id
		}
	}
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if aiKey := os.Getenv("AI_API_KEY"); aiKey != "" {
		cfg.AI.APIKey = aiKey
		cfg.AI.Enabled = true
	}
	if aiBase := os.Getenv("AI_BASE_URL"); aiBase != "" {
		cfg.AI.BaseURL = aiBase
	}
	if aiModel := os.Getenv("AI_MODEL"); aiModel != "" {
		cfg.AI.Model = aiModel
	}

	if len(cfg.Labels) == 0 {
		cfg.Labels = []string{"frontend", "backend", "auth", "ui", "api", "db", "devops"}
	}

	if cfg.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram.token is required (set in %s or TELEGRAM_TOKEN env)", path)
	}

	return cfg, nil
}
