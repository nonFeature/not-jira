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
	ID_CHECK = "5251750898467649056" // ✅ Done / Success / Approved
	ID_CROSS = "5251563676548246141" // ❌ Rejected / Error
	ID_BUG = "5253671487583333045" // 🪲 Bug
	ID_IDEA = "5251566481161888804" // 💡 Idea
	ID_NEW = "5251224984017217415" // 🆕 New task status
	ID_GEAR = "5251569208466121931" // ⚙️ In progress / Settings
	ID_WAVE = "5251642639521981799" // 👋 Greeting
	ID_STAR = "5251402275972228214" // ⭐️ Admin section
	ID_CLIPBOARD = "5251696537066578863" // 📋 Tasks list / Card header
	ID_MEMO = "5251544155921883411" // 📝 Description / Form
	ID_PENCIL = "5251450895002016858" // ✏️ Edit title
	ID_PLUS = "5251309384419548860" // ➕ Add subtask
	ID_MESSAGES = "5251355271850141254" // 💬 Comment
	ID_LINK = "5251669332743725645" // 🔗 Original post link
	ID_REFRESH = "5251394437656911929" // 🔄 Refresh button
	ID_ARROW_L = "5253747813447150577" // ⬅️ Prev page
	ID_ARROW_R = "5251477790087225246" // ➡️ Next page
	ID_ROCKET = "5251203324497145220" // 🚀 Write in DM button
	ID_BELL = "5251588119207126938" // 🔔 Notifications enabled
	ID_BELL_OFF = "5251335016784373065" // 🔕 Notifications disabled
	ID_INFO = "5251750791093464353" // ℹ️ Usage / Information
	ID_WARNING = "5251531176530714957" // ⚠️ Warning
	ID_LOCK = "5251400021114398284" // ⛔️ Access denied (Admin only)
	ID_HOURGLASS = "5253819444911710732" // ⏳ Processing AI
	ID_PARTY = "5251214796354804294" // 🎉 Task created successfully
	ID_SUB_EMPTY = "5251582771972844119" // ⬜️ Subtask unchecked
	ID_SUB_DONE = "5251391302330787578" // ☑️ Subtask checked
	ID_BOX = "5251483382134646909" // 📦 Archive / Box
	ID_USER = "5253956862390345938" // 👤 Assignee / Person
	ID_TAG = "5253876499257271563" // 🏷️ Label / Tag
	ID_PRIORITY = "5253449291745241072" // ⚡️ Priority button
	ID_P0 = "5251665716381262011" // 🔴 Blocker / P0
	ID_P1 = "5253657614838965029" // 🟡 High / P1
	ID_P2 = "5251552303474843852" // 🔵 Normal / P2
	ID_P3 = "5251467138568333349" // ⚪️ Low / P3
	ID_TRASH = "5253830659071322395" // 🗑️ Delete / Clear / Trash
	ID_WRITE = "5251444916407542521" // ✍️ Created by me tab in /my
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
func Box() string { return E(ID_BOX, "📦") }
func User() string { return E(ID_USER, "👤") }
func Tag() string { return E(ID_TAG, "🏷️") }
func Priority() string { return E(ID_PRIORITY, "⚡️") }
func P0() string { return E(ID_P0, "🔴") }
func P1() string { return E(ID_P1, "🟡") }
func P2() string { return E(ID_P2, "🔵") }
func P3() string { return E(ID_P3, "⚪️") }
func Trash() string { return E(ID_TRASH, "🗑️") }
func Write() string { return E(ID_WRITE, "✍️") }

// FallbackForID returns standard Unicode emoji for given custom emoji ID.
func FallbackForID(emojiID string) string {
	if emojiID == "" {
		return ""
	}
	switch emojiID {
	case ID_CHECK:
		return "✅"
	case ID_CROSS:
		return "❌"
	case ID_BUG:
		return "🪲"
	case ID_IDEA:
		return "💡"
	case ID_NEW:
		return "🆕"
	case ID_GEAR:
		return "⚙️"
	case ID_WAVE:
		return "👋"
	case ID_STAR:
		return "⭐️"
	case ID_CLIPBOARD:
		return "📋"
	case ID_MEMO:
		return "📝"
	case ID_PENCIL:
		return "✏️"
	case ID_PLUS:
		return "➕"
	case ID_MESSAGES:
		return "💬"
	case ID_LINK:
		return "🔗"
	case ID_REFRESH:
		return "🔄"
	case ID_ARROW_L:
		return "⬅️"
	case ID_ARROW_R:
		return "➡️"
	case ID_ROCKET:
		return "🚀"
	case ID_BELL:
		return "🔔"
	case ID_BELL_OFF:
		return "🔕"
	case ID_INFO:
		return "ℹ️"
	case ID_WARNING:
		return "⚠️"
	case ID_LOCK:
		return "⛔️"
	case ID_HOURGLASS:
		return "⏳"
	case ID_PARTY:
		return "🎉"
	case ID_SUB_EMPTY:
		return "⬜️"
	case ID_SUB_DONE:
		return "☑️"
	case ID_BOX:
		return "📦"
	case ID_USER:
		return "👤"
	case ID_TAG:
		return "🏷️"
	case ID_PRIORITY:
		return "⚡️"
	case ID_P0:
		return "🔴"
	case ID_P1:
		return "🟡"
	case ID_P2:
		return "🔵"
	case ID_P3:
		return "⚪️"
	case ID_TRASH:
		return "🗑"
	case ID_WRITE:
		return "✍️"
	default:
		return ""
	}
}

// MakeInlineButton constructs a telego.InlineKeyboardButton with icon_custom_emoji_id or fallback.
// It guarantees that text never contains raw <tg-emoji> tags.
func MakeInlineButton(text, callbackData, url, emojiID, fallbackEmoji, style string) telego.InlineKeyboardButton {
	cleanText := StripCustomEmojis(text)
	btn := telego.InlineKeyboardButton{
		CallbackData: callbackData,
		URL:          url,
		Style:        style,
	}

	if useCustomEmoji.Load() && emojiID != "" {
		if cleanText == "" {
			cleanText = " "
		}
		btn.Text = cleanText
		btn.IconCustomEmojiID = emojiID
	} else {
		if fallbackEmoji != "" {
			if cleanText != "" {
				btn.Text = fallbackEmoji + " " + cleanText
			} else {
				btn.Text = fallbackEmoji
			}
		} else {
			btn.Text = cleanText
		}
	}

	return btn
}

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
		strings.Contains(msg, "button_custom_emoji_invalid") ||
		strings.Contains(msg, "custom_emoji")
}

// CheckAndUpdateFromMessage inspects the sent Telegram message entities and inline keyboard.
// If the sent message requested custom emojis but Telegram stripped them (no custom_emoji entity or button icon),
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
	if !hasCustom && msg.ReplyMarkup != nil && len(msg.ReplyMarkup.InlineKeyboard) > 0 {
		for _, row := range msg.ReplyMarkup.InlineKeyboard {
			for _, btn := range row {
				if btn.IconCustomEmojiID != "" {
					hasCustom = true
					break
				}
			}
			if hasCustom {
				break
			}
		}
	}

	if !hasCustom && useCustomEmoji.Load() {
		log.Println("[Emoji] Telegram stripped custom_emoji from sent message or buttons. Disabling custom emojis and falling back to unicode.")
		useCustomEmoji.Store(false)
	}

	return hasCustom
}
