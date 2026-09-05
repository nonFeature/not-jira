package emoji

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/mymmrac/telego"
)

var (
	useCustomEmoji atomic.Bool
	tgEmojiRegex   = regexp.MustCompile(`<\/?tg-emoji[^>]*>`)
)

func init() {
	useCustomEmoji.Store(true)
}

// SetCustomEmojiEnabled manually sets custom emojis state
func SetCustomEmojiEnabled(enabled bool) {
	useCustomEmoji.Store(enabled)
}

// IsCustomEmojiEnabled returns whether custom emojis are enabled
func IsCustomEmojiEnabled() bool {
	return useCustomEmoji.Load()
}

// =========================================================================
// CUSTOM EMOJI IDs
// Paste your custom emoji IDs here from your Telegram custom emoji pack.
// When left empty (""), the bot uses standard Unicode fallback emojis.
// =========================================================================
var (
	ID_CHECK = "" // ✅ Done / Success / Approved
	ID_CROSS = "" // ❌ Rejected / Error
	ID_BUG = "" // 🪲 Bug
	ID_IDEA = "" // 💡 Idea
	ID_NEW = "" // 🆕 New task status
	ID_GEAR = "" // ⚙️ In progress / Settings
	ID_WAVE = "" // 👋 Greeting
	ID_STAR = "" // ⭐️ Admin section
	ID_CLIPBOARD = "" // 📋 Tasks list / Card header
	ID_MEMO = "" // 📝 Description / Form
	ID_PENCIL = "" // ✏️ Edit title
	ID_PLUS = "" // ➕ Add subtask
	ID_MESSAGES = "" // 💬 Comment
	ID_LINK = "" // 🔗 Original post link
	ID_REFRESH = "" // 🔄 Refresh button
	ID_ARROW_L = "" // ⬅️ Prev page
	ID_ARROW_R = "" // ➡️ Next page
	ID_ROCKET = "" // 🚀 Write in DM button
	ID_BELL = "" // 🔔 Notifications enabled
	ID_BELL_OFF = "" // 🔕 Notifications disabled
	ID_INFO = "" // ℹ️ Usage / Information
	ID_WARNING = "" // ⚠️ Warning
	ID_LOCK = "" // ⛔️ Access denied (Admin only)
	ID_HOURGLASS = "" // ⏳ Processing AI
	ID_PARTY = "" // 🎉 Task created successfully
	ID_SUB_EMPTY = "" // ⬜️ Subtask unchecked
	ID_SUB_DONE = "" // ☑️ Subtask checked
)

// E wraps a fallback emoji into <tg-emoji emoji-id="..."> tag if custom emojis are enabled and ID is set.
func E(emojiID string, fallback string) string {
	if useCustomEmoji.Load() && emojiID != "" {
		return fmt.Sprintf(`<tg-emoji emoji-id="%s">%s</tg-emoji>`, emojiID, fallback)
	}
	return fallback
}

// Emoji Accessors (mimicking PluginsBot _EmojiAccessor)
func Check() string { return E(ID_CHECK, "✅") }
func Cross() string { return E(ID_CROSS, "❌") }
func Bug() string { return E(ID_BUG, "🪲") }
func Idea() string { return E(ID_IDEA, "💡") }
func New() string { return E(ID_NEW, "🆕") }
func Gear() string { return E(ID_GEAR, "⚙️") }
func Wave() string { return E(ID_WAVE, "👋") }
func Star() string { return E(ID_STAR, "⭐️") }
func Clipboard() string { return E(ID_CLIPBOARD, "📋") }
func Memo() string { return E(ID_MEMO, "📝") }
func Pencil() string { return E(ID_PENCIL, "✏️") }
func Plus() string { return E(ID_PLUS, "➕") }
func Messages() string { return E(ID_MESSAGES, "💬") }
func Link() string { return E(ID_LINK, "🔗") }
func Refresh() string { return E(ID_REFRESH, "🔄") }
func ArrowL() string { return E(ID_ARROW_L, "⬅️") }
func ArrowR() string { return E(ID_ARROW_R, "➡️") }
func Rocket() string { return E(ID_ROCKET, "🚀") }
func Bell() string { return E(ID_BELL, "🔔") }
func BellOff() string { return E(ID_BELL_OFF, "🔕") }
func Info() string { return E(ID_INFO, "ℹ️") }
func Warning() string { return E(ID_WARNING, "⚠️") }
func Lock() string { return E(ID_LOCK, "⛔️") }
func Hourglass() string { return E(ID_HOURGLASS, "⏳") }
func Party() string { return E(ID_PARTY, "🎉") }
func SubEmpty() string { return E(ID_SUB_EMPTY, "⬜️") }
func SubDone() string { return E(ID_SUB_DONE, "☑️") }

// StripCustomEmojis removes all <tg-emoji> and </tg-emoji> tags, leaving the fallback emoji inside.
func StripCustomEmojis(text string) string {
	return tgEmojiRegex.ReplaceAllString(text, "")
}

// IsCustomEmojiError checks if Telegram rejected the message because of custom emojis.
func IsCustomEmojiError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "custom emoji") ||
		strings.Contains(msg, "entity_custom_emoji_invalid") ||
		strings.Contains(msg, "custom_emoji")
}

// CheckAndUpdateFromMessage inspects the sent Telegram message entities.
// If the sent message requested custom emojis but Telegram stripped them (no custom_emoji entity),
// it automatically sets useCustomEmoji to false.
func CheckAndUpdateFromMessage(msg *telego.Message, hadCustomEmoji bool) bool {
	if msg == nil || !hadCustomEmoji {
		return useCustomEmoji.Load()
	}

	hasCustom := false
	for _, ent := range msg.Entities {
		if ent.Type == "custom_emoji" {
			hasCustom = true
			break
		}
	}
	if !hasCustom {
		for _, ent := range msg.CaptionEntities {
			if ent.Type == "custom_emoji" {
				hasCustom = true
				break
			}
		}
	}

	if !hasCustom && useCustomEmoji.Load() {
		log.Println("[Emoji] Telegram stripped custom_emoji entities from sent message. Disabling custom emojis and falling back to unicode.")
		useCustomEmoji.Store(false)
	}

	return hasCustom
}
