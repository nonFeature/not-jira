package bot

import (
	"context"
	"fmt"

	"not-jira/internal/config"
	"not-jira/internal/locales"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type UserHandler struct {
	bot         *telego.Bot
	botUsername string
	cfg         *config.Config
	storage     storage.Storage
}

func NewUserHandler(bot *telego.Bot, botUsername string, cfg *config.Config, storage storage.Storage) *UserHandler {
	return &UserHandler{
		bot:         bot,
		botUsername: botUsername,
		cfg:         cfg,
		storage:     storage,
	}
}

func (h *UserHandler) HandleStart(ctx context.Context, msg *telego.Message) {
	l := locales.ForUser(msg.From.LanguageCode)
	isAdmin := h.cfg.Telegram.IsAdmin(msg.From.ID)

	text := l.Start.GreetingUser
	if isAdmin {
		text += l.Start.GreetingAdmin
	}

	_, err := SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(msg.From.ID), text).WithParseMode(telego.ModeHTML))
	if err != nil && msg.Chat.ID != msg.From.ID {
		PromptStartInDM(ctx, h.bot, h.botUsername, msg)
	}
}

func (h *UserHandler) HandleSettings(ctx context.Context, msg *telego.Message) {
	l := locales.ForUser(msg.From.LanguageCode)
	settings, err := h.storage.GetUserSettings(ctx, msg.From.ID)
	if err != nil {
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(msg.From.ID), "❌ Error loading settings."))
		return
	}

	stateDesc := l.Settings.StatusEnabled
	if !settings.NotifyDM {
		stateDesc = l.Settings.StatusDisabled
	}

	text := fmt.Sprintf(l.Settings.Title, stateDesc)
	kb := BuildSettingsKeyboard(settings.NotifyDM, l)

	_, err = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(msg.From.ID), text).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb))
	if err != nil && msg.Chat.ID != msg.From.ID {
		PromptStartInDM(ctx, h.bot, h.botUsername, msg)
	}
}

func (h *UserHandler) HandleToggleNotify(ctx context.Context, query *telego.CallbackQuery) {
	l := locales.ForUser(query.From.LanguageCode)
	userID := query.From.ID
	settings, err := h.storage.GetUserSettings(ctx, userID)
	if err != nil {
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Error"))
		return
	}

	newStatus := !settings.NotifyDM
	if query.Data == "settings:enable" {
		newStatus = true
	} else if query.Data == "settings:disable" {
		newStatus = false
	}

	if err := h.storage.SetNotifyDM(ctx, userID, newStatus); err != nil {
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Save error"))
		return
	}

	stateDesc := l.Settings.StatusEnabled
	if !newStatus {
		stateDesc = l.Settings.StatusDisabled
	}

	text := fmt.Sprintf(l.Settings.Title, stateDesc)
	kb := BuildSettingsKeyboard(newStatus, l)

	_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText(l.Settings.UpdatedAlert))

	editParams := &telego.EditMessageTextParams{
		ChatID:      tu.ID(query.Message.GetChat().ID),
		MessageID:   query.Message.GetMessageID(),
		Text:        text,
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: kb,
	}
	_, _ = EditMessageTextSafe(ctx, h.bot, editParams)
}
