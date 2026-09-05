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
		{"ru", "not-jira"},
		{"ru-RU", "not-jira"},
		{"RU", "not-jira"},
		{"en", "not-jira"},
		{"en-US", "not-jira"},
		{"de", "not-jira"},
		{"", "not-jira"},
	}

	for _, tt := range tests {
		bundle := ForUser(tt.lang)
		if !strings.Contains(bundle.Start.GreetingUser, tt.expected) {
			t.Errorf("for lang %q expected substring %q, got %q", tt.lang, tt.expected, bundle.Start.GreetingUser)
		}
	}
}
