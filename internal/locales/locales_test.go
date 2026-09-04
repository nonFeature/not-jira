package locales

import "testing"

func TestForUser(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		{"ru", "👋 <b>Привет! Я not-jira.</b>\n\n"},
		{"ru-RU", "👋 <b>Привет! Я not-jira.</b>\n\n"},
		{"RU", "👋 <b>Привет! Я not-jira.</b>\n\n"},
		{"en", "👋 <b>Hello! I am not-jira.</b>\n\n"},
		{"en-US", "👋 <b>Hello! I am not-jira.</b>\n\n"},
		{"de", "👋 <b>Hello! I am not-jira.</b>\n\n"},
		{"", "👋 <b>Hello! I am not-jira.</b>\n\n"},
	}

	for _, tt := range tests {
		bundle := ForUser(tt.lang)
		if len(bundle.Start.GreetingUser) < 30 || bundle.Start.GreetingUser[:30] != tt.expected[:30] {
			t.Errorf("for lang %q expected prefix %q, got %q", tt.lang, tt.expected[:30], bundle.Start.GreetingUser[:30])
		}
	}
}
