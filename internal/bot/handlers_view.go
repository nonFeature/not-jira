package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"not-jira/internal/config"
	"not-jira/internal/locales"
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
	l := locales.ForUser(msg.From.LanguageCode)
	h.renderList(ctx, targetChatID, 0, "ALL", "ALL", "ALL", 0, msg, l)
}

func (h *ViewHandler) HandleView(ctx context.Context, msg *telego.Message) {
	// Always send task card strictly to the user's private messages
	targetChatID := msg.From.ID
	l := locales.ForUser(msg.From.LanguageCode)
	parts := strings.Fields(msg.Text)
	if len(parts) < 2 {
		reply := tu.Message(tu.ID(targetChatID), l.View.UsageHint).WithParseMode(telego.ModeHTML)
		_, err := SendMessageSafe(ctx, h.bot, reply)
		if err != nil && msg.Chat.ID != msg.From.ID {
			PromptStartInDM(ctx, h.bot, h.botUsername, msg)
		}
		return
	}

	taskID := strings.ToUpper(strings.TrimSpace(parts[1]))
	h.renderTask(ctx, targetChatID, 0, taskID, msg.From.ID, msg, l)
}

func (h *ViewHandler) HandleMyTasks(ctx context.Context, msg *telego.Message) {
	// Always send personal tasks strictly to the user's private messages
	targetChatID := msg.From.ID
	l := locales.ForUser(msg.From.LanguageCode)
	h.renderMyTasks(ctx, targetChatID, 0, msg.From.ID, "assigned", 0, msg, l)
}

func (h *ViewHandler) HandleCallback(ctx context.Context, query *telego.CallbackQuery) {
	data := query.Data
	l := locales.ForUser(query.From.LanguageCode)

	if data == "noop" {
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "list:") {
		// Can be list:{type}:{status}:{page} (4 parts) or list:{type}:{status}:{tag}:{page} (5 parts)
		parts := strings.Split(data, ":")
		var filterType, filterStatus, filterTag string
		var page int
		if len(parts) == 4 {
			filterType = parts[1]
			filterStatus = parts[2]
			filterTag = "ALL"
			page, _ = strconv.Atoi(parts[3])
		} else if len(parts) >= 5 {
			filterType = parts[1]
			filterStatus = parts[2]
			filterTag = parts[3]
			page, _ = strconv.Atoi(parts[4])
		}
		if filterType != "" && filterStatus != "" {
			chatID := query.Message.GetChat().ID
			msgID := query.Message.GetMessageID()
			h.renderList(ctx, chatID, msgID, filterType, filterStatus, filterTag, page, nil, l)
		}
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "list_tags:") {
		// list_tags:{type}:{status}:{tag}:{page}
		parts := strings.Split(data, ":")
		if len(parts) >= 5 {
			filterType := parts[1]
			filterStatus := parts[2]
			filterTag := parts[3]
			page, _ := strconv.Atoi(parts[4])
			chatID := query.Message.GetChat().ID
			msgID := query.Message.GetMessageID()

			availableTags, _ := h.storage.GetAllLabels(ctx)
			kb := BuildListTagsKeyboard(filterType, filterStatus, filterTag, page, availableTags, l)
			editMsg := &telego.EditMessageReplyMarkupParams{
				ChatID:      tu.ID(chatID),
				MessageID:   msgID,
				ReplyMarkup: kb,
			}
			_, _ = h.bot.EditMessageReplyMarkup(ctx, editMsg)
		}
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "my:") {
		// my:{tab}:{page}
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			tab := parts[1]
			page, _ := strconv.Atoi(parts[2])
			chatID := query.Message.GetChat().ID
			msgID := query.Message.GetMessageID()
			h.renderMyTasks(ctx, chatID, msgID, query.From.ID, tab, page, nil, l)
		}
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "view:") {
		// view:{task_id}
		taskID := strings.TrimPrefix(data, "view:")
		chatID := query.Message.GetChat().ID
		msgID := query.Message.GetMessageID()
		h.renderTask(ctx, chatID, msgID, taskID, query.From.ID, nil, l)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}
}

func (h *ViewHandler) renderList(ctx context.Context, chatID int64, editMsgID int, filterType, filterStatus, filterTag string, page int, originMsg *telego.Message, l *locales.Bundle) {
	var filter storage.TaskFilter
	if filterType != "ALL" {
		t := models.TaskType(filterType)
		filter.Type = &t
	}
	if filterStatus != "ALL" {
		s := models.TaskStatus(filterStatus)
		filter.Status = &s
	}
	if filterTag != "" && filterTag != "ALL" {
		filter.Label = filterTag
	}

	offset := page * PageSize
	tasks, totalCount, err := h.storage.ListTasks(ctx, filter, offset, PageSize)
	if err != nil {
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(chatID), fmt.Sprintf("❌ Error: %v", err)))
		return
	}

	totalPages := (totalCount + PageSize - 1) / PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	header := RenderTaskListHeader(totalCount, filterType, filterStatus, filterTag, l)
	kb := BuildListKeyboard(tasks, filterType, filterStatus, filterTag, page, totalPages, l)

	if editMsgID != 0 {
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(chatID),
			MessageID:   editMsgID,
			Text:        header,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, err = EditMessageTextSafe(ctx, h.bot, editMsg)
		if err == nil {
			return
		}
	}

	msg := tu.Message(tu.ID(chatID), header).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb)
	_, err = SendMessageSafe(ctx, h.bot, msg)
	if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
		PromptStartInDM(ctx, h.bot, h.botUsername, originMsg)
	}
}

func (h *ViewHandler) renderTask(ctx context.Context, chatID int64, editMsgID int, taskID string, userID int64, originMsg *telego.Message, l *locales.Bundle) {
	task, err := h.storage.GetTask(ctx, taskID)
	if err != nil || task == nil {
		_, err := SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(chatID), fmt.Sprintf(l.View.NotFound, taskID)).WithParseMode(telego.ModeHTML))
		if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
			PromptStartInDM(ctx, h.bot, h.botUsername, originMsg)
		}
		return
	}

	isAdmin := h.cfg.Telegram.IsAdmin(userID)
	isDev := h.cfg.Telegram.IsDev(userID)
	cardHTML := RenderTaskCard(task, l)
	kb := BuildTaskInlineKeyboard(task, userID, isAdmin, isDev, l)

	if editMsgID != 0 {
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(chatID),
			MessageID:   editMsgID,
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, err = EditMessageTextSafe(ctx, h.bot, editMsg)
		if err == nil {
			return
		}
	}

	msg := tu.Message(tu.ID(chatID), cardHTML).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb)
	_, err = SendMessageSafe(ctx, h.bot, msg)
	if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
		PromptStartInDM(ctx, h.bot, h.botUsername, originMsg)
	}
}

func (h *ViewHandler) renderMyTasks(ctx context.Context, chatID int64, editMsgID int, userID int64, tab string, page int, originMsg *telego.Message, l *locales.Bundle) {
	var filter storage.TaskFilter
	var tabName string
	if tab == "created" {
		filter.AuthorID = &userID
		tabName = l.Filters.MyCreated
	} else {
		tab = "assigned"
		filter.AssigneeID = &userID
		tabName = l.Filters.MyAssigned
	}

	offset := page * PageSize
	tasks, totalCount, err := h.storage.ListTasks(ctx, filter, offset, PageSize)
	if err != nil {
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(chatID), fmt.Sprintf("❌ Error: %v", err)))
		return
	}

	totalPages := (totalCount + PageSize - 1) / PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	header := fmt.Sprintf(l.View.MyTasksHeader, totalCount, tabName)
	if totalCount == 0 {
		header += "<i>" + l.View.NoMyTasks + "</i>\n"
	}

	kb := BuildMyTasksKeyboard(tasks, tab, page, totalPages, l)

	if editMsgID != 0 {
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(chatID),
			MessageID:   editMsgID,
			Text:        header,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, err = EditMessageTextSafe(ctx, h.bot, editMsg)
		if err == nil {
			return
		}
	}

	msg := tu.Message(tu.ID(chatID), header).WithParseMode(telego.ModeHTML).WithReplyMarkup(kb)
	_, err = SendMessageSafe(ctx, h.bot, msg)
	if err != nil && originMsg != nil && originMsg.Chat.ID != originMsg.From.ID {
		PromptStartInDM(ctx, h.bot, h.botUsername, originMsg)
	}
}
