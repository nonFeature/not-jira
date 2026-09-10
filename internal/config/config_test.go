package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("telegram:\n  token: \"from-file\"\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	t.Setenv("TELEGRAM_TOKEN", "from-env")
	t.Setenv("TELEGRAM_FORUM_CHAT_ID", "-1001234567890")
	t.Setenv("TELEGRAM_FORUM_URL", "https://t.me/+test")
	t.Setenv("AI_API_KEY", "test-key")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Telegram.Token != "from-env" {
		t.Errorf("expected env token override, got %q", cfg.Telegram.Token)
	}
	if cfg.Telegram.ForumChatID != -1001234567890 {
		t.Errorf("expected forum chat ID override, got %d", cfg.Telegram.ForumChatID)
	}
	if cfg.Telegram.ForumURL != "https://t.me/+test" {
		t.Errorf("expected forum URL override, got %q", cfg.Telegram.ForumURL)
	}
	if !cfg.AI.Enabled || cfg.AI.APIKey != "test-key" {
		t.Errorf("expected AI enabled by env key, got enabled=%v key=%q", cfg.AI.Enabled, cfg.AI.APIKey)
	}
}

func TestLoadMissingToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("telegram:\n  proxy_url: \"\"\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	t.Setenv("TELEGRAM_TOKEN", "")

	if _, err := Load(path); err == nil {
		t.Fatalf("expected error when telegram token is missing")
	}
}
