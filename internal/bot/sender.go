package bot

import (
	"context"
	"log"
	"strings"

	"not-jira/internal/emoji"

	"github.com/mymmrac/telego"
)

// StripKeyboardCustomEmojis removes custom emojis from an inline keyboard and prefixes text with fallback.
func StripKeyboardCustomEmojis(ikm *telego.InlineKeyboardMarkup) {
	if ikm == nil {
		return
	}
	for i := range ikm.InlineKeyboard {
		for j := range ikm.InlineKeyboard[i] {
			btn := &ikm.InlineKeyboard[i][j]
			btn.Text = emoji.StripCustomEmojis(btn.Text)
			if btn.IconCustomEmojiID != "" {
				fb := emoji.FallbackForID(btn.IconCustomEmojiID)
				if fb != "" && !strings.HasPrefix(btn.Text, fb) {
					btn.Text = fb + " " + btn.Text
				}
				btn.IconCustomEmojiID = ""
			}
		}
	}
}

func stripReplyMarkup(markup telego.ReplyMarkup) telego.ReplyMarkup {
	if markup == nil {
		return nil
	}
	if ikm, ok := markup.(*telego.InlineKeyboardMarkup); ok && ikm != nil {
		StripKeyboardCustomEmojis(ikm)
	}
	return markup
}

func keyboardHadCustomEmoji(markup telego.ReplyMarkup) bool {
	if ikm, ok := markup.(*telego.InlineKeyboardMarkup); ok && ikm != nil {
		for _, row := range ikm.InlineKeyboard {
			for _, btn := range row {
				if btn.IconCustomEmojiID != "" {
					return true
				}
			}
		}
	}
	return false
}

// SendMessageSafe sends a message. If custom emojis are rejected or disabled, it strips them and resends.
func SendMessageSafe(ctx context.Context, bot *telego.Bot, params *telego.SendMessageParams) (*telego.Message, error) {
	if params == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	hadCustomEmoji := strings.Contains(params.Text, "<tg-emoji") || keyboardHadCustomEmoji(params.ReplyMarkup)
	if strings.Contains(params.Text, "<tg-emoji") && params.ParseMode == "" {
		params.ParseMode = telego.ModeHTML
	}
	if !emoji.IsCustomEmojiEnabled() {
		if strings.Contains(params.Text, "<tg-emoji") {
			params.Text = emoji.StripCustomEmojis(params.Text)
		}
		stripReplyMarkup(params.ReplyMarkup)
		hadCustomEmoji = false
	}

	msg, err := bot.SendMessage(ctx, params)
	if err != nil && emoji.IsCustomEmojiError(err) {
		log.Printf("[Bot WARNING] Telegram API rejected custom emoji (%v). Falling back to unicode emojis.", err)
		emoji.SetCustomEmojiEnabled(false)
		params.Text = emoji.StripCustomEmojis(params.Text)
		stripReplyMarkup(params.ReplyMarkup)
		return bot.SendMessage(ctx, params)
	}

	if err == nil && msg != nil {
		emoji.CheckAndUpdateFromMessage(msg, hadCustomEmoji)
	}

	return msg, err
}

// EditMessageTextSafe edits a message text with custom emoji fallback.
func EditMessageTextSafe(ctx context.Context, bot *telego.Bot, params *telego.EditMessageTextParams) (*telego.Message, error) {
	if params == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	hadCustomEmoji := strings.Contains(params.Text, "<tg-emoji") || keyboardHadCustomEmoji(params.ReplyMarkup)
	if strings.Contains(params.Text, "<tg-emoji") && params.ParseMode == "" {
		params.ParseMode = telego.ModeHTML
	}
	if !emoji.IsCustomEmojiEnabled() {
		if strings.Contains(params.Text, "<tg-emoji") {
			params.Text = emoji.StripCustomEmojis(params.Text)
		}
		StripKeyboardCustomEmojis(params.ReplyMarkup)
		hadCustomEmoji = false
	}

	msg, err := bot.EditMessageText(ctx, params)
	if err != nil && emoji.IsCustomEmojiError(err) {
		log.Printf("[Bot WARNING] Telegram API rejected custom emoji on edit (%v). Falling back to unicode emojis.", err)
		emoji.SetCustomEmojiEnabled(false)
		params.Text = emoji.StripCustomEmojis(params.Text)
		StripKeyboardCustomEmojis(params.ReplyMarkup)
		return bot.EditMessageText(ctx, params)
	}

	if err == nil && msg != nil {
		emoji.CheckAndUpdateFromMessage(msg, hadCustomEmoji)
	}

	return msg, err
}
