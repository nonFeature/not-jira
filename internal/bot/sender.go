package bot

import (
	"log"
	"strings"

	"not-jira/internal/emoji"

	"github.com/mymmrac/telego"
)

// SendMessageSafe sends a message. If custom emojis are rejected or disabled, it strips them and resends.
func SendMessageSafe(bot *telego.Bot, params *telego.SendMessageParams) (*telego.Message, error) {
	if params == nil {
		return nil, nil
	}

	hadCustomEmoji := strings.Contains(params.Text, "<tg-emoji")
	if !emoji.IsCustomEmojiEnabled() && hadCustomEmoji {
		params.Text = emoji.StripCustomEmojis(params.Text)
		hadCustomEmoji = false
	}

	msg, err := bot.SendMessage(params)
	if err != nil && emoji.IsCustomEmojiError(err) {
		log.Printf("[Bot WARNING] Telegram API rejected custom emoji (%v). Falling back to unicode emojis.", err)
		emoji.SetCustomEmojiEnabled(false)
		params.Text = emoji.StripCustomEmojis(params.Text)
		return bot.SendMessage(params)
	}

	if err == nil && msg != nil {
		emoji.CheckAndUpdateFromMessage(msg, hadCustomEmoji)
	}

	return msg, err
}

// EditMessageTextSafe edits a message text with custom emoji fallback.
func EditMessageTextSafe(bot *telego.Bot, params *telego.EditMessageTextParams) (*telego.Message, error) {
	if params == nil {
		return nil, nil
	}

	hadCustomEmoji := strings.Contains(params.Text, "<tg-emoji")
	if !emoji.IsCustomEmojiEnabled() && hadCustomEmoji {
		params.Text = emoji.StripCustomEmojis(params.Text)
		hadCustomEmoji = false
	}

	msg, err := bot.EditMessageText(params)
	if err != nil && emoji.IsCustomEmojiError(err) {
		log.Printf("[Bot WARNING] Telegram API rejected custom emoji on edit (%v). Falling back to unicode emojis.", err)
		emoji.SetCustomEmojiEnabled(false)
		params.Text = emoji.StripCustomEmojis(params.Text)
		return bot.EditMessageText(params)
	}

	if err == nil && msg != nil {
		emoji.CheckAndUpdateFromMessage(msg, hadCustomEmoji)
	}

	return msg, err
}
