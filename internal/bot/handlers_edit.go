package bot

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"not-jira/internal/config"
	"not-jira/internal/emoji"
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

		taskID := parts[1]
		newStatus := models.TaskStatus(parts[2])

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		oldStatus := task.Status
		task.Status = newStatus
		if err := h.storage.UpdateTask(ctx, task); err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Status update error"))
			return
		}

		h.answerAlert(ctx, query.ID, fmt.Sprintf(l.Edit.StatusChangedAlert, TaskStatusUnicode(newStatus), TaskStatusName(newStatus, l)), false)

		// Notify topic & author if status changed
		if oldStatus != newStatus {
			h.notifier.NotifyStatusChange(ctx, task)
		}

		// Update the current message card
		h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		return
	}

	if strings.HasPrefix(data, "reopen:") {
		taskID := strings.TrimPrefix(data, "reopen:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		task.IsArchived = false
		task.Status = models.StatusInProgress
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, l.Edit.StatusReopenedAlert, false)
		h.notifier.NotifyStatusChange(ctx, task)
		h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		return
	}

	if strings.HasPrefix(data, "archive:") {
		taskID := strings.TrimPrefix(data, "archive:")
		if !isAdmin {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		task.IsArchived = true
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, l.Edit.TaskArchivedAlert, false)
		h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		return
	}

	if strings.HasPrefix(data, "claim:") {
		taskID := strings.TrimPrefix(data, "claim:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !isAdmin && !h.cfg.Telegram.IsDev(userID) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		task.AssigneeID = userID
		task.AssigneeUsername = query.From.Username
		if task.AssigneeUsername == "" {
			task.AssigneeUsername = query.From.FirstName
		}
		if task.Status == models.StatusNew {
			task.Status = models.StatusInProgress
		}
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, fmt.Sprintf(l.Edit.TaskClaimedNotify, task.AssigneeUsername, task.ID), false)
		h.notifier.NotifyStatusChange(ctx, task)
		h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		return
	}

	if strings.HasPrefix(data, "transfer:") {
		taskID := strings.TrimPrefix(data, "transfer:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAssigningTask,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptTransfer, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "accept_assign:") {
		taskID := strings.TrimPrefix(data, "accept_assign:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		task.AssigneeID = userID
		task.AssigneeUsername = query.From.Username
		if task.AssigneeUsername == "" {
			task.AssigneeUsername = query.From.FirstName
		}
		if task.Status == models.StatusNew {
			task.Status = models.StatusInProgress
		}
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, fmt.Sprintf(l.Edit.TransferAcceptedNotify, task.AssigneeUsername, task.ID), false)
		_, _ = EditMessageTextSafe(ctx, h.bot, &telego.EditMessageTextParams{
			ChatID:    tu.ID(userID),
			MessageID: query.Message.GetMessageID(),
			Text:      fmt.Sprintf(l.Edit.TransferAcceptedNotify, task.AssigneeUsername, task.ID),
			ParseMode: telego.ModeHTML,
		})
		h.notifier.NotifyStatusChange(ctx, task)
		return
	}

	if strings.HasPrefix(data, "reject_assign:") {
		taskID := strings.TrimPrefix(data, "reject_assign:")
		username := query.From.Username
		if username == "" {
			username = query.From.FirstName
		}
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = EditMessageTextSafe(ctx, h.bot, &telego.EditMessageTextParams{
			ChatID:    tu.ID(userID),
			MessageID: query.Message.GetMessageID(),
			Text:      fmt.Sprintf(l.Edit.TransferRejectedNotify, username, taskID),
			ParseMode: telego.ModeHTML,
		})
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

		task, _ := h.storage.GetTask(ctx, taskID)
		if task != nil && !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		_, err := h.storage.ToggleSubtask(ctx, subID)
		if err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Toggle error"))
			return
		}

		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

		task, _ = h.storage.GetTask(ctx, taskID)
		if task != nil {
			h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		}
		return
	}

	// GitHub Issues edit actions
	if strings.HasPrefix(data, "edit_title:") {
		taskID := strings.TrimPrefix(data, "edit_title:")
		if !isAdmin {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateEditingTitle,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditTitle, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "edit_desc:") {
		taskID := strings.TrimPrefix(data, "edit_desc:")
		if !isAdmin {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateEditingDesc,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditDesc, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "add_sub:") {
		taskID := strings.TrimPrefix(data, "add_sub:")
		task, _ := h.storage.GetTask(ctx, taskID)
		if task != nil && !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAddingSubtask,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddSubtask, taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "add_comm:") {
		taskID := strings.TrimPrefix(data, "add_comm:")
		task, _ := h.storage.GetTask(ctx, taskID)
		if task != nil && !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAddingComment,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddComment, taskID)).WithParseMode(telego.ModeHTML))
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
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.Cancelled))
		return true
	}

	switch sess.State {
	case models.StateCreatingTaskTitle:
		sess.DraftTask.Title = text
		sess.State = models.StateCreatingTaskDesc
		h.fsm.Set(userID, sess)

		prompt := l.Add.FormDescPrompt
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), prompt).WithParseMode(telego.ModeHTML))
		return true

	case models.StateCreatingTaskDesc:
		if text != "-" {
			sess.DraftTask.Description = text
		}

		if err := h.storage.CreateTask(ctx, sess.DraftTask); err != nil {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf("❌ Error: %v", err)))
			h.fsm.Clear(userID)
			return true
		}

		task := sess.DraftTask
		h.fsm.Clear(userID)

		cardHTML := RenderTaskCard(task, l)
		kb := BuildTaskInlineKeyboard(task, userID, true, h.cfg.Telegram.IsDev(userID), l)

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
			_, _ = SendMessageSafe(ctx, h.bot, topicMsg)
		}

		// 2. Send management card to admin in DM
		confirmDM := tu.Message(tu.ID(userID), l.Add.FormCreatedSuccess+cardHTML).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(kb)
		_, _ = SendMessageSafe(ctx, h.bot, confirmDM)
		return true

	case models.StateEditingTitle:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Title = text
			_ = h.storage.UpdateTask(ctx, task)
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TitleUpdated, task.ID, html.EscapeString(text))).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingDesc:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Description = text
			_ = h.storage.UpdateTask(ctx, task)
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.DescUpdated, task.ID)).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAddingSubtask:
		_, err := h.storage.AddSubtask(ctx, sess.TaskID, text)
		if err == nil {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.SubtaskAdded, sess.TaskID, html.EscapeString(text))).WithParseMode(telego.ModeHTML))
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
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.CommentAdded, sess.TaskID)).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAssigningTask:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err != nil || task == nil {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.View.NotFound, sess.TaskID)))
			h.fsm.Clear(userID)
			return true
		}

		var targetUID int64
		var targetUsername string

		// Forwarded message?
		if msg.ForwardOrigin != nil {
			if userOrigin, ok := msg.ForwardOrigin.(*telego.MessageOriginUser); ok && userOrigin != nil {
				targetUID = userOrigin.SenderUser.ID
				targetUsername = userOrigin.SenderUser.Username
				if targetUsername == "" {
					targetUsername = userOrigin.SenderUser.FirstName
				}
			}
		}

		if targetUID == 0 {
			cleanText := strings.TrimPrefix(strings.TrimSpace(text), "@")
			if id, err := strconv.ParseInt(cleanText, 10, 64); err == nil && id > 0 {
				targetUID = id
				targetUsername = cleanText
			} else {
				foundID, _ := h.storage.FindUserIDByUsername(ctx, cleanText)
				if foundID != 0 {
					targetUID = foundID
					targetUsername = cleanText
				}
			}
		}

		if targetUID == 0 {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TransferUserNotFound, text)))
			h.fsm.Clear(userID)
			return true
		}

		senderName := msg.From.Username
		if senderName == "" {
			senderName = msg.From.FirstName
		}

		// Send offer DM to target user
		offerText := fmt.Sprintf(l.Edit.TransferOfferReceived, senderName, task.ID, html.EscapeString(task.Title), html.EscapeString(task.Description))
		inviteKb := BuildTransferInviteKeyboard(task.ID, l)
		_, err = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(targetUID), offerText).WithParseMode(telego.ModeHTML).WithReplyMarkup(inviteKb))
		if err != nil {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TransferUserNotFound, targetUsername)))
		} else {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TransferOfferSent, targetUsername)))
		}
		h.fsm.Clear(userID)
		return true
	}

	return false
}

func (h *EditHandler) updateMessageCard(ctx context.Context, chatID int64, messageID int, task *models.Task, userID int64, l *locales.Bundle) {
	cardHTML := RenderTaskCard(task, l)
	isAdmin := h.cfg.Telegram.IsAdmin(userID)
	isDev := h.cfg.Telegram.IsDev(userID)
	kb := BuildTaskInlineKeyboard(task, userID, isAdmin, isDev, l)

	editParams := &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		Text:        cardHTML,
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: kb,
	}
	_, _ = EditMessageTextSafe(ctx, h.bot, editParams)
}

func (h *EditHandler) answerAlert(ctx context.Context, queryID string, text string, showAlert bool) {
	cb := tu.CallbackQuery(queryID).WithText(cleanAlertText(text))
	if showAlert {
		cb = cb.WithShowAlert()
	}
	_ = h.bot.AnswerCallbackQuery(ctx, cb)
}

func cleanAlertText(s string) string {
	clean := emoji.StripCustomEmojis(s)
	clean = strings.ReplaceAll(clean, "<b>", "")
	clean = strings.ReplaceAll(clean, "</b>", "")
	clean = strings.ReplaceAll(clean, "<i>", "")
	clean = strings.ReplaceAll(clean, "</i>", "")
	clean = strings.ReplaceAll(clean, "<code>", "")
	clean = strings.ReplaceAll(clean, "</code>", "")
	return clean
}

