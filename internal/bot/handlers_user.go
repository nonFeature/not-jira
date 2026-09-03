package bot

import (
	"context"
	"fmt"

	"not-jira/internal/config"
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
	isAdmin := h.cfg.Telegram.IsAdmin(msg.From.ID)

	text := "👋 <b>Привет! Я not-jira.</b>\n\n" +
		"Я помогаю собирать и отслеживать задачи, баги и идеи прямо в топиках форума.\n\n" +
		"<b>Команды:</b>\n" +
		"• <code>/list</code> — Просмотр списка задач с фильтрами и страницами\n" +
		"• <code>/view [ID]</code> — Просмотр карточки задачи (например, <code>/view B0</code>)\n" +
		"• <code>/settings</code> — Настройки уведомлений в ЛС\n" +
		"• <code>/help</code> — Справка\n\n"

	if isAdmin {
		text += "⭐️ <b>Команды администратора:</b>\n" +
			"• <code>/add</code> — Ответьте на сообщение в топике форума, чтобы зарегистрировать баг (B) или идею (I)\n" +
			"• В карточке задачи вам доступны кнопки изменения статуса, заголовка, описания, саб-тасков и комментариев.\n"
	}

	_, err := h.bot.SendMessage(tu.Message(tu.ID(msg.From.ID), text).WithParseMode(telego.ModeHTML))
	if err != nil && msg.Chat.ID != msg.From.ID {
		PromptStartInDM(h.bot, h.botUsername, msg)
	}
}

func (h *UserHandler) HandleSettings(ctx context.Context, msg *telego.Message) {
	settings, err := h.storage.GetUserSettings(ctx, msg.From.ID)
	if err != nil {
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(msg.From.ID), "❌ Ошибка загрузки настроек."))
		return
	}

	stateDesc := "включены 🔔"
	if !settings.NotifyDM {
		stateDesc = "отключены 🔕"
	}

	text := fmt.Sprintf("⚙️ <b>Настройки уведомлений</b>\n\nУведомления в личные сообщения: <b>%s</b>\n\nВы можете изменить настройку нажатием кнопки:", stateDesc)
	kb := BuildSettingsKeyboard(settings.NotifyDM)

	_, err = h.bot.SendMessage(tu.Message(tu.ID(msg.From.ID), text).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb))
	if err != nil && msg.Chat.ID != msg.From.ID {
		PromptStartInDM(h.bot, h.botUsername, msg)
	}
}

func (h *UserHandler) HandleToggleNotify(ctx context.Context, query *telego.CallbackQuery) {
	userID := query.From.ID
	settings, err := h.storage.GetUserSettings(ctx, userID)
	if err != nil {
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Ошибка"))
		return
	}

	newStatus := !settings.NotifyDM
	if err := h.storage.SetNotifyDM(ctx, userID, newStatus); err != nil {
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Ошибка сохранения"))
		return
	}

	stateDesc := "включены 🔔"
	if !newStatus {
		stateDesc = "отключены 🔕"
	}

	text := fmt.Sprintf("⚙️ <b>Настройки уведомлений</b>\n\nУведомления в личные сообщения: <b>%s</b>\n\nВы можете изменить настройку нажатием кнопки:", stateDesc)
	kb := BuildSettingsKeyboard(newStatus)

	_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("Настройки обновлены"))

	editParams := &telego.EditMessageTextParams{
		ChatID:      tu.ID(query.Message.GetChat().ID),
		MessageID:   query.Message.GetMessageID(),
		Text:        text,
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: kb,
	}
	_, _ = h.bot.EditMessageText(editParams)
}
