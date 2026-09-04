package bot

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"not-jira/internal/config"
	"not-jira/internal/locales"
	"not-jira/internal/models"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type EditHandler struct {
	bot      *telego.Bot
	cfg      *config.Config
	storage  storage.Storage
	fsm      *FSM
	notifier *Notifier
}

func NewEditHandler(bot *telego.Bot, cfg *config.Config, storage storage.Storage, fsm *FSM, notifier *Notifier) *EditHandler {
	return &EditHandler{
		bot:      bot,
		cfg:      cfg,
		storage:  storage,
		fsm:      fsm,
		notifier: notifier,
	}
}

func (h *EditHandler) HandleCallback(ctx context.Context, query *telego.CallbackQuery) {
	data := query.Data
	userID := query.From.ID
	l := locales.ForUser(query.From.LanguageCode)
	isAdmin := h.cfg.Telegram.IsAdmin(userID)

	if strings.HasPrefix(data, "set_status:") {
		// set_status:{task_id}:{status}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		if !isAdmin {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText(l.Common.AdminOnly).WithShowAlert())
			return
		}

		taskID := parts[1]
		newStatus := models.TaskStatus(parts[2])

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText(fmt.Sprintf(l.View.NotFound, taskID)))
			return
		}

		oldStatus := task.Status
		task.Status = newStatus
		if err := h.storage.UpdateTask(ctx, task); err != nil {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Status update error"))
			return
		}

		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText(fmt.Sprintf(l.Edit.StatusChangedAlert, TaskStatusEmoji(newStatus), TaskStatusName(newStatus, l))))

		// Notify topic & author if status changed
		if oldStatus != newStatus {
			h.notifier.NotifyStatusChange(ctx, task)
		}

		// Update the current message card
		h.updateMessageCard(query.Message.GetChat().ID, query.Message.GetMessageID(), task, isAdmin, l)
		return
	}

	if strings.HasPrefix(data, "toggle_sub:") {
		// toggle_sub:{subtask_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		subID, _ := strconv.ParseInt(parts[1], 10, 64)
		taskID := parts[2]

		_, err := h.storage.ToggleSubtask(ctx, subID)
		if err != nil {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Toggle error"))
			return
		}

		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))

		task, _ := h.storage.GetTask(ctx, taskID)
		if task != nil {
			h.updateMessageCard(query.Message.GetChat().ID, query.Message.GetMessageID(), task, isAdmin, l)
		}
		return
	}

	// GitHub Issues edit actions
	if !isAdmin {
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText(l.Common.AdminOnly).WithShowAlert())
		return
	}

	if strings.HasPrefix(data, "edit_title:") {
		taskID := strings.TrimPrefix(data, "edit_title:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateEditingTitle,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditTitle, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "edit_desc:") {
		taskID := strings.TrimPrefix(data, "edit_desc:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateEditingDesc,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditDesc, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "add_sub:") {
		taskID := strings.TrimPrefix(data, "add_sub:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAddingSubtask,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddSubtask, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "add_comm:") {
		taskID := strings.TrimPrefix(data, "add_comm:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAddingComment,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddComment, taskID)).WithParseMode(telego.ModeHTML))
		return
	}
}

func (h *EditHandler) HandleFSMMessage(ctx context.Context, msg *telego.Message) bool {
	userID := msg.From.ID
	l := locales.ForUser(msg.From.LanguageCode)
	sess := h.fsm.Get(userID)
	if sess == nil || sess.State == models.StateNone {
		return false
	}

	text := strings.TrimSpace(msg.Text)
	if text == "/cancel" {
		h.fsm.Clear(userID)
		_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), l.Common.Cancelled))
		return true
	}

	switch sess.State {
	case models.StateCreatingTaskTitle:
		sess.DraftTask.Title = text
		sess.State = models.StateCreatingTaskDesc
		h.fsm.Set(userID, sess)

		prompt := l.Add.FormDescPrompt
		_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), prompt).WithParseMode(telego.ModeHTML))
		return true

	case models.StateCreatingTaskDesc:
		if text != "-" {
			sess.DraftTask.Description = text
		}

		if err := h.storage.CreateTask(ctx, sess.DraftTask); err != nil {
			_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf("❌ Error: %v", err)))
			h.fsm.Clear(userID)
			return true
		}

		task := sess.DraftTask
		h.fsm.Clear(userID)

		cardHTML := RenderTaskCard(task, l)
		kb := BuildTaskInlineKeyboard(task, true, l)

		// 1. If origin chat was a group/topic, send brief confirmation to user in topic
		if task.ChatID != userID {
			var topicReplyText string
			if task.Type == models.TaskTypeIdea {
				topicReplyText = fmt.Sprintf(l.Add.AcceptedIdea, task.ID)
			} else {
				topicReplyText = fmt.Sprintf(l.Add.AcceptedBug, task.ID)
			}
			topicMsg := tu.Message(tu.ID(task.ChatID), topicReplyText).WithParseMode(telego.ModeHTML)
			if task.MessageID != 0 {
				topicMsg.ReplyParameters = &telego.ReplyParameters{MessageID: int(task.MessageID)}
			}
			if task.TopicID != 0 {
				topicMsg.MessageThreadID = int(task.TopicID)
			}
			_, _ = SendMessageSafe(h.bot, topicMsg)
		}

		// 2. Send management card to admin in DM
		confirmDM := tu.Message(tu.ID(userID), l.Add.FormCreatedSuccess+cardHTML).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(kb)
		_, _ = SendMessageSafe(h.bot, confirmDM)
		return true

	case models.StateEditingTitle:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Title = text
			_ = h.storage.UpdateTask(ctx, task)
			_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TitleUpdated, task.ID, html.EscapeString(text))).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingDesc:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Description = text
			_ = h.storage.UpdateTask(ctx, task)
			_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.DescUpdated, task.ID)).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAddingSubtask:
		_, err := h.storage.AddSubtask(ctx, sess.TaskID, text)
		if err == nil {
			_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.SubtaskAdded, sess.TaskID, html.EscapeString(text))).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAddingComment:
		authorName := msg.From.Username
		if authorName == "" {
			authorName = msg.From.FirstName
		}
		_, err := h.storage.AddComment(ctx, sess.TaskID, userID, authorName, text)
		if err == nil {
			_, _ = SendMessageSafe(h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.CommentAdded, sess.TaskID)).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true
	}

	return false
}

func (h *EditHandler) updateMessageCard(chatID int64, messageID int, task *models.Task, isAdmin bool, l *locales.Bundle) {
	cardHTML := RenderTaskCard(task, l)
	kb := BuildTaskInlineKeyboard(task, isAdmin, l)

	editParams := &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		Text:        cardHTML,
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: kb,
	}
	_, _ = EditMessageTextSafe(h.bot, editParams)
}
