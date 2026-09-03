package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"not-jira/internal/config"
	"not-jira/internal/models"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const PageSize = 5

type ViewHandler struct {
	bot         *telego.Bot
	botUsername string
	cfg         *config.Config
	storage     storage.Storage
}

func NewViewHandler(bot *telego.Bot, botUsername string, cfg *config.Config, storage storage.Storage) *ViewHandler {
	return &ViewHandler{
		bot:         bot,
		botUsername: botUsername,
		cfg:         cfg,
		storage:     storage,
	}
}

func (h *ViewHandler) HandleList(ctx context.Context, msg *telego.Message) {
	// Always send the list strictly to the user's private messages
	targetChatID := msg.From.ID
	h.renderList(ctx, targetChatID, 0, "ALL", "ALL", 0, msg)
}

func (h *ViewHandler) HandleView(ctx context.Context, msg *telego.Message) {
	// Always send task card strictly to the user's private messages
	targetChatID := msg.From.ID
	parts := strings.Fields(msg.Text)
	if len(parts) < 2 {
		reply := tu.Message(tu.ID(targetChatID), "ℹ️ Использование: <code>/view B0</code> или <code>/view I1</code>").WithParseMode(telego.ModeHTML)
		_, err := h.bot.SendMessage(reply)
		if err != nil && msg.Chat.ID != msg.From.ID {
			PromptStartInDM(h.bot, h.botUsername, msg)
		}
		return
	}

	taskID := strings.ToUpper(strings.TrimSpace(parts[1]))
	h.renderTask(ctx, targetChatID, 0, taskID, msg.From.ID, msg)
}

func (h *ViewHandler) HandleCallback(ctx context.Context, query *telego.CallbackQuery) {
	data := query.Data

	if data == "noop" {
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "list:") {
		// list:{type}:{status}:{page}
		parts := strings.Split(data, ":")
		if len(parts) == 4 {
			filterType := parts[1]
			filterStatus := parts[2]
			page, _ := strconv.Atoi(parts[3])
			chatID := query.Message.GetChat().ID
			msgID := query.Message.GetMessageID()
			h.renderList(ctx, chatID, msgID, filterType, filterStatus, page, nil)
		}
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "view:") {
		// view:{task_id}
		taskID := strings.TrimPrefix(data, "view:")
		chatID := query.Message.GetChat().ID
		msgID := query.Message.GetMessageID()
		h.renderTask(ctx, chatID, msgID, taskID, query.From.ID, nil)
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		return
	}
}

func (h *ViewHandler) renderList(ctx context.Context, chatID int64, editMsgID int, filterType, filterStatus string, page int, originMsg *telego.Message) {
	var filter storage.TaskFilter
	if filterType != "ALL" {
		t := models.TaskType(filterType)
		filter.Type = &t
	}
	if filterStatus != "ALL" {
		s := models.TaskStatus(filterStatus)
		filter.Status = &s
	}

	offset := page * PageSize
	tasks, totalCount, err := h.storage.ListTasks(ctx, filter, offset, PageSize)
	if err != nil {
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(chatID), fmt.Sprintf("❌ Ошибка получения задач: %v", err)))
		return
	}

	totalPages := (totalCount + PageSize - 1) / PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	header := RenderTaskListHeader(totalCount, filterType, filterStatus)
	kb := BuildListKeyboard(tasks, filterType, filterStatus, page, totalPages)

	if editMsgID != 0 {
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(chatID),
			MessageID:   editMsgID,
			Text:        header,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, err = h.bot.EditMessageText(editMsg)
		if err == nil {
			return
		}
	}

	msg := tu.Message(tu.ID(chatID), header).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb)
	_, err = h.bot.SendMessage(msg)
	if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
		PromptStartInDM(h.bot, h.botUsername, originMsg)
	}
}

func (h *ViewHandler) renderTask(ctx context.Context, chatID int64, editMsgID int, taskID string, userID int64, originMsg *telego.Message) {
	task, err := h.storage.GetTask(ctx, taskID)
	if err != nil || task == nil {
		_, err := h.bot.SendMessage(tu.Message(tu.ID(chatID), fmt.Sprintf("❌ Задача <code>%s</code> не найдена.", taskID)).WithParseMode(telego.ModeHTML))
		if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
			PromptStartInDM(h.bot, h.botUsername, originMsg)
		}
		return
	}

	isAdmin := h.cfg.Telegram.IsAdmin(userID)
	cardHTML := RenderTaskCard(task)
	kb := BuildTaskInlineKeyboard(task, isAdmin)

	if editMsgID != 0 {
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(chatID),
			MessageID:   editMsgID,
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, err = h.bot.EditMessageText(editMsg)
		if err == nil {
			return
		}
	}

	msg := tu.Message(tu.ID(chatID), cardHTML).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb)
	_, err = h.bot.SendMessage(msg)
	if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
		PromptStartInDM(h.bot, h.botUsername, originMsg)
	}
}
