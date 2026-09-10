package bot

import (
	"context"
	"log"
	"strings"

	"not-jira/internal/models"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
)

func userDisplayName(u *telego.User) string {
	if u == nil {
		return ""
	}
	name := strings.TrimSpace(u.Username)
	if name == "" {
		name = strings.TrimSpace(u.FirstName)
	}
	return name
}

func recordHistory(ctx context.Context, st storage.Storage, taskID string, user *telego.User, action, oldValue, newValue string) {
	if st == nil || taskID == "" {
		return
	}

	entry := &models.HistoryEntry{
		TaskID:   taskID,
		Action:   action,
		OldValue: oldValue,
		NewValue: newValue,
	}
	if user != nil {
		entry.AuthorID = user.ID
		entry.AuthorName = userDisplayName(user)
	}

	if err := st.AddHistory(ctx, entry); err != nil {
		log.Printf("[History WARN] Failed to record %q for %s: %v", action, taskID, err)
	}
}
