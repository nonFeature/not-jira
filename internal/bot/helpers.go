package bot

import (
	"context"
	"fmt"

	"not-jira/internal/emoji"
	"not-jira/internal/locales"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// PromptStartInDM sends a reply in the group/topic asking the user to start the bot in private chat.
// Automatically adapts language to user's LanguageCode (Russian or English).
func PromptStartInDM(ctx context.Context, bot *telego.Bot, botUsername string, originMsg *telego.Message) {
	if originMsg == nil || originMsg.Chat.ID == originMsg.From.ID {
		return
	}

	l := locales.ForUser(originMsg.From.LanguageCode)

	userMention := originMsg.From.FirstName
	if originMsg.From.Username != "" {
		userMention = "@" + originMsg.From.Username
	}

	text := fmt.Sprintf(l.Notifications.PromptStartDM, userMention)
	reply := tu.Message(tu.ID(originMsg.Chat.ID), text)
	reply.ReplyParameters = &telego.ReplyParameters{MessageID: originMsg.MessageID}
	if originMsg.MessageThreadID != 0 {
		reply.MessageThreadID = originMsg.MessageThreadID
	}

	if botUsername != "" {
		btn := emoji.MakeInlineButton(l.Buttons.OpenDM, "", fmt.Sprintf("https://t.me/%s", botUsername), emoji.ID_ROCKET, "🚀", "primary")
		reply.ReplyMarkup = &telego.InlineKeyboardMarkup{
			InlineKeyboard: [][]telego.InlineKeyboardButton{
				{btn},
			},
		}
	}

	_, _ = SendMessageSafe(ctx, bot, reply)
}
