package bot

import (
	"context"
	"fmt"
	"html"
	"log"
	"sync"
	"time"

	"not-jira/internal/locales"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const memberCacheTTL = 5 * time.Minute

type memberCacheEntry struct {
	isMember bool
	expires  time.Time
}

type memberCache struct {
	mu      sync.Mutex
	entries map[int64]memberCacheEntry
	ttl     time.Duration
}

func newMemberCache() *memberCache {
	return &memberCache{
		entries: make(map[int64]memberCacheEntry),
		ttl:     memberCacheTTL,
	}
}

func (c *memberCache) get(userID int64) (bool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[userID]
	if !ok || time.Now().After(entry.expires) {
		return false, false
	}
	return entry.isMember, true
}

func (c *memberCache) set(userID int64, isMember bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[userID] = memberCacheEntry{
		isMember: isMember,
		expires:  time.Now().Add(c.ttl),
	}
}

func (c *memberCache) cleanup(maxAge time.Duration) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0
	for uid, entry := range c.entries {
		if entry.expires.Before(cutoff) {
			delete(c.entries, uid)
			cleaned++
		}
	}
	return cleaned
}

func (s *BotService) isForumMember(ctx context.Context, userID int64) bool {
	if s.cfg.Telegram.ForumChatID == 0 || s.cfg.Telegram.IsDev(userID) {
		return true
	}

	if isMember, ok := s.memberCache.get(userID); ok {
		return isMember
	}

	member, err := s.bot.GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: tu.ID(s.cfg.Telegram.ForumChatID),
		UserID: userID,
	})
	if err != nil {
		log.Printf("[Membership WARN] Failed to check membership of user %d: %v", userID, err)
		return true
	}

	isMember := member.MemberIsMember()
	s.memberCache.set(userID, isMember)
	return isMember
}

func (s *BotService) rejectNonMember(ctx context.Context, update telego.Update, userID int64) {
	lang := ""
	if update.Message != nil && update.Message.From != nil {
		lang = update.Message.From.LanguageCode
	} else if update.CallbackQuery != nil {
		lang = update.CallbackQuery.From.LanguageCode
	}
	l := locales.ForUser(lang)

	text := l.Common.ForumRequired
	if s.cfg.Telegram.ForumURL != "" {
		text += fmt.Sprintf(l.Common.JoinForumLink, html.EscapeString(s.cfg.Telegram.ForumURL))
	}

	if update.CallbackQuery != nil {
		alertText := cleanAlertText(l.Common.ForumRequired)
		if s.cfg.Telegram.ForumURL != "" {
			alertText += "\n" + s.cfg.Telegram.ForumURL
		}
		_ = s.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(update.CallbackQuery.ID).WithText(alertText).WithShowAlert())
		return
	}

	msg := tu.Message(tu.ID(userID), text).WithParseMode(telego.ModeHTML)
	msg.LinkPreviewOptions = &telego.LinkPreviewOptions{IsDisabled: true}
	_, _ = SendMessageSafe(ctx, s.bot, msg)
}
