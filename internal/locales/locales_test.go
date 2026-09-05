package locales

import (
	"strings"
	"testing"
)

func TestForUser(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		{"ru", "<b>not-jira</b>"},
		{"ru-RU", "<b>not-jira</b>"},
		{"RU", "<b>not-jira</b>"},
		{"en", "<b>not-jira</b>"},
		{"en-US", "<b>not-jira</b>"},
		{"de", "<b>not-jira</b>"},
		{"", "<b>not-jira</b>"},
	}

	for _, tt := range tests {
		bundle := ForUser(tt.lang)
		if !strings.Contains(bundle.Start.GreetingUser, tt.expected) {
			t.Errorf("for lang %q expected substring %q, got %q", tt.lang, tt.expected, bundle.Start.GreetingUser)
		}
	}
}
