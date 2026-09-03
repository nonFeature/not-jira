package bot

import (
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// PromptStartInDM sends a reply in the group/topic asking the user to start the bot in private chat
func PromptStartInDM(bot *telego.Bot, botUsername string, originMsg *telego.Message) {
	if originMsg == nil || originMsg.Chat.ID == originMsg.From.ID {
		return
	}

	userMention := originMsg.From.FirstName
	if originMsg.From.Username != "" {
		userMention = "@" + originMsg.From.Username
	}

	text := fmt.Sprintf("⚠️ %s, чтобы бот мог отвечать вам, сначала начните с ним диалог в личных сообщениях.", userMention)
	reply := tu.Message(tu.ID(originMsg.Chat.ID), text)
	reply.ReplyParameters = &telego.ReplyParameters{MessageID: originMsg.MessageID}
	if originMsg.MessageThreadID != 0 {
		reply.MessageThreadID = originMsg.MessageThreadID
	}

	if botUsername != "" {
		reply.ReplyMarkup = &telego.InlineKeyboardMarkup{
			InlineKeyboard: [][]telego.InlineKeyboardButton{
				{
					{Text: "🚀 Написать боту в ЛС", URL: fmt.Sprintf("https://t.me/%s", botUsername)},
				},
			},
		}
	}

	_, _ = bot.SendMessage(reply)
}
