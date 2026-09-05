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

	if strings.HasPrefix(data, "unclaim:") {
		taskID := strings.TrimPrefix(data, "unclaim:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if task.AssigneeID != userID && !isAdmin {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		task.AssigneeID = 0
		task.AssigneeUsername = ""
		if task.Status == models.StatusInProgress {
			task.Status = models.StatusNew
		}
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, l.Edit.TaskUnclaimedAlert, false)
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
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptTransfer, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
		return
	}

	if strings.HasPrefix(data, "accept_assign:") {
		// accept_assign:{task_id} or accept_assign:{task_id}:{target_uid}
		parts := strings.Split(data, ":")
		if len(parts) < 2 {
			return
		}
		taskID := parts[1]
		if len(parts) >= 3 {
			targetUID, err := strconv.ParseInt(parts[2], 10, 64)
			if err == nil && targetUID != 0 && targetUID != userID {
				h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
				return
			}
		}

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
		// reject_assign:{task_id} or reject_assign:{task_id}:{target_uid}
		parts := strings.Split(data, ":")
		if len(parts) < 2 {
			return
		}
		taskID := parts[1]
		if len(parts) >= 3 {
			targetUID, err := strconv.ParseInt(parts[2], 10, 64)
			if err == nil && targetUID != 0 && targetUID != userID {
				h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
				return
			}
		}

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

	// Cancel FSM button
	if strings.HasPrefix(data, "cancel_fsm") {
		taskID := strings.TrimPrefix(data, "cancel_fsm:")
		if taskID == "cancel_fsm" {
			taskID = ""
		}
		h.fsm.Clear(userID)
		h.answerAlert(ctx, query.ID, l.Common.Cancelled, false)
		if taskID != "" {
			task, err := h.storage.GetTask(ctx, taskID)
			if err == nil && task != nil {
				h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
			}
		} else {
			editMsg := &telego.EditMessageTextParams{
				ChatID:    tu.ID(query.Message.GetChat().ID),
				MessageID: query.Message.GetMessageID(),
				Text:      l.Common.Cancelled,
			}
			_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
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
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditTitle, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
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
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditDesc, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
		return
	}

	if strings.HasPrefix(data, "manage_sub:") || strings.HasPrefix(data, "add_sub:") {
		taskID := strings.TrimPrefix(data, "manage_sub:")
		taskID = strings.TrimPrefix(taskID, "add_sub:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		if len(task.Subtasks) == 0 {
			h.fsm.Set(userID, &models.UserSession{
				State:  models.StateAddingSubtask,
				TaskID: taskID,
			})
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddSubtask, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
			return
		}

		cardHTML := RenderTaskCard(task, l)
		kb := BuildSubtasksManageKeyboard(task, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "sub_item:") {
		// sub_item:{subtask_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		subID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		taskID := parts[2]

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		var targetSub *models.Subtask
		itemNum := 0
		for i := range task.Subtasks {
			if task.Subtasks[i].ID == subID {
				targetSub = &task.Subtasks[i]
				itemNum = i + 1
				break
			}
		}
		if targetSub == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		cardHTML := RenderTaskCard(task, l)
		kb := BuildSubtaskItemKeyboard(task.ID, targetSub, itemNum, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "add_sub_prompt:") {
		taskID := strings.TrimPrefix(data, "add_sub_prompt:")
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
			State:  models.StateAddingSubtask,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddSubtask, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
		return
	}

	if strings.HasPrefix(data, "edit_sub:") {
		// edit_sub:{subtask_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		subID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		taskID := parts[2]

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
			State:     models.StateEditingSubtask,
			TaskID:    taskID,
			SubtaskID: subID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditSubtask, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
		return
	}

	if strings.HasPrefix(data, "del_sub:") {
		// del_sub:{subtask_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		subID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		taskID := parts[2]

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		if err := h.storage.DeleteSubtask(ctx, subID); err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Delete error"))
			return
		}

		h.answerAlert(ctx, query.ID, l.Edit.SubtaskDeletedAlert, false)

		task, _ = h.storage.GetTask(ctx, taskID)
		if task != nil {
			cardHTML := RenderTaskCard(task, l)
			var kb *telego.InlineKeyboardMarkup
			if len(task.Subtasks) > 0 {
				kb = BuildSubtasksManageKeyboard(task, l)
			} else {
				kb = BuildTaskInlineKeyboard(task, userID, isAdmin, h.cfg.Telegram.IsDev(userID), l)
			}
			editMsg := &telego.EditMessageTextParams{
				ChatID:      tu.ID(query.Message.GetChat().ID),
				MessageID:   query.Message.GetMessageID(),
				Text:        cardHTML,
				ParseMode:   telego.ModeHTML,
				ReplyMarkup: kb,
			}
			_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		}
		return
	}

	if strings.HasPrefix(data, "clear_subs:") {
		taskID := strings.TrimPrefix(data, "clear_subs:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		if err := h.storage.ClearSubtasks(ctx, taskID); err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Clear error"))
			return
		}

		h.answerAlert(ctx, query.ID, l.Edit.SubtasksClearedAlert, false)

		task, _ = h.storage.GetTask(ctx, taskID)
		if task != nil {
			h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		}
		return
	}

	if strings.HasPrefix(data, "manage_comm:") || strings.HasPrefix(data, "add_comm:") {
		taskID := strings.TrimPrefix(data, "manage_comm:")
		taskID = strings.TrimPrefix(taskID, "add_comm:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		if len(task.Comments) == 0 {
			h.fsm.Set(userID, &models.UserSession{
				State:  models.StateAddingComment,
				TaskID: taskID,
			})
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddComment, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
			return
		}

		cardHTML := RenderTaskCard(task, l)
		kb := BuildCommentsManageKeyboard(task, userID, isAdmin, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "comm_item:") {
		// comm_item:{comment_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		commID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		taskID := parts[2]

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		var targetComment *models.Comment
		itemNum := 0
		for i := range task.Comments {
			if task.Comments[i].ID == commID {
				targetComment = &task.Comments[i]
				itemNum = i + 1
				break
			}
		}
		if targetComment == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		canEdit := isAdmin || (targetComment.AuthorID != 0 && targetComment.AuthorID == userID)
		if !canEdit {
			h.answerAlert(ctx, query.ID, l.Edit.CannotEditOtherComment, true)
			return
		}

		cardHTML := RenderTaskCard(task, l)
		kb := BuildCommentItemKeyboard(task.ID, targetComment, itemNum, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "add_comm_prompt:") {
		taskID := strings.TrimPrefix(data, "add_comm_prompt:")
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
			State:  models.StateAddingComment,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptAddComment, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
		return
	}

	if strings.HasPrefix(data, "edit_comm:") {
		// edit_comm:{comment_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		commID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		taskID := parts[2]

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		var targetComment *models.Comment
		for i := range task.Comments {
			if task.Comments[i].ID == commID {
				targetComment = &task.Comments[i]
				break
			}
		}
		if targetComment == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		if !isAdmin && (targetComment.AuthorID == 0 || targetComment.AuthorID != userID) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		h.fsm.Set(userID, &models.UserSession{
			State:     models.StateEditingComment,
			TaskID:    taskID,
			CommentID: commID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditComment, taskID)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
		return
	}

	if strings.HasPrefix(data, "del_comm:") {
		// del_comm:{comment_id}:{task_id}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		commID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		taskID := parts[2]

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		var targetComment *models.Comment
		for i := range task.Comments {
			if task.Comments[i].ID == commID {
				targetComment = &task.Comments[i]
				break
			}
		}
		if targetComment == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		if !isAdmin && (targetComment.AuthorID == 0 || targetComment.AuthorID != userID) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		if err := h.storage.DeleteComment(ctx, commID); err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Delete error"))
			return
		}

		h.answerAlert(ctx, query.ID, l.Edit.CommentDeletedAlert, false)

		task, _ = h.storage.GetTask(ctx, taskID)
		if task != nil {
			cardHTML := RenderTaskCard(task, l)
			var kb *telego.InlineKeyboardMarkup
			if len(task.Comments) > 0 {
				kb = BuildCommentsManageKeyboard(task, userID, isAdmin, l)
			} else {
				kb = BuildTaskInlineKeyboard(task, userID, isAdmin, h.cfg.Telegram.IsDev(userID), l)
			}
			editMsg := &telego.EditMessageTextParams{
				ChatID:      tu.ID(query.Message.GetChat().ID),
				MessageID:   query.Message.GetMessageID(),
				Text:        cardHTML,
				ParseMode:   telego.ModeHTML,
				ReplyMarkup: kb,
			}
			_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		}
		return
	}

	if strings.HasPrefix(data, "clear_comms:") {
		taskID := strings.TrimPrefix(data, "clear_comms:")
		if !isAdmin {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}

		if err := h.storage.ClearComments(ctx, taskID); err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Clear error"))
			return
		}

		h.answerAlert(ctx, query.ID, l.Edit.CommentsClearedAlert, false)

		task, _ = h.storage.GetTask(ctx, taskID)
		if task != nil {
			h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		}
		return
	}

	if strings.HasPrefix(data, "pick_priority:") {
		taskID := strings.TrimPrefix(data, "pick_priority:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		cardHTML := RenderTaskCard(task, l)
		kb := BuildPriorityPickerKeyboard(task.ID, task.Priority, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "set_priority:") {
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		taskID := parts[1]
		p := models.TaskPriority(parts[2])

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		task.Priority = p
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, fmt.Sprintf(l.Edit.PriorityChangedAlert, task.Priority.Emoji(), TaskPriorityName(task.Priority, l)), false)
		h.updateMessageCard(ctx, query.Message.GetChat().ID, query.Message.GetMessageID(), task, userID, l)
		return
	}

	if strings.HasPrefix(data, "edit_labels:") || strings.HasPrefix(data, "pick_labels:") {
		taskID := strings.TrimPrefix(data, "edit_labels:")
		taskID = strings.TrimPrefix(taskID, "pick_labels:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		availableTags := h.getAvailableTags(ctx, task)
		cardHTML := RenderTaskCard(task, l)
		kb := BuildTaskLabelsKeyboard(task, availableTags, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		return
	}

	if strings.HasPrefix(data, "toggle_label:") {
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		taskID := parts[1]
		tag := parts[2]
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		added := task.ToggleLabel(tag)
		_ = h.storage.UpdateTask(ctx, task)

		alertMsg := fmt.Sprintf(l.Edit.LabelToggledOnAlert, tag)
		if !added {
			alertMsg = fmt.Sprintf(l.Edit.LabelToggledOffAlert, tag)
		}
		h.answerAlert(ctx, query.ID, alertMsg, false)

		availableTags := h.getAvailableTags(ctx, task)
		cardHTML := RenderTaskCard(task, l)
		kb := BuildTaskLabelsKeyboard(task, availableTags, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		h.notifier.NotifyStatusChange(ctx, task)
		return
	}

	if strings.HasPrefix(data, "clear_labels:") {
		taskID := strings.TrimPrefix(data, "clear_labels:")
		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			h.answerAlert(ctx, query.ID, fmt.Sprintf(l.View.NotFound, taskID), false)
			return
		}
		if !task.CanManage(userID, isAdmin) {
			h.answerAlert(ctx, query.ID, l.Common.AdminOnly, true)
			return
		}

		task.Labels = nil
		_ = h.storage.UpdateTask(ctx, task)

		h.answerAlert(ctx, query.ID, l.Edit.LabelsClearedAlert, false)

		availableTags := h.getAvailableTags(ctx, task)
		cardHTML := RenderTaskCard(task, l)
		kb := BuildTaskLabelsKeyboard(task, availableTags, l)
		editMsg := &telego.EditMessageTextParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			Text:        cardHTML,
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: kb,
		}
		_, _ = EditMessageTextSafe(ctx, h.bot, editMsg)
		h.notifier.NotifyStatusChange(ctx, task)
		return
	}

	if strings.HasPrefix(data, "edit_labels_manual:") {
		taskID := strings.TrimPrefix(data, "edit_labels_manual:")
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
			State:  models.StateEditingLabels,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

		var tagList []string
		for _, t := range h.getAvailableTags(ctx, task) {
			tagList = append(tagList, "#"+t)
		}
		tagsFormatted := strings.Join(tagList, "  ")
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.PromptEditLabels, taskID, tagsFormatted)).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(taskID, l)))
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
	if text == "" && msg.Caption != "" {
		text = strings.TrimSpace(msg.Caption)
	}


	if text == "" && sess.State != models.StateAssigningTask {
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
		return true
	}

	switch sess.State {
	case models.StateCreatingTaskTitle:
		cleanTitle := strings.Join(strings.Fields(strings.ReplaceAll(text, "\n", " ")), " ")
		cleanTitle = emoji.StripCustomEmojis(cleanTitle)
		if cleanTitle == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard("", l)))
			return true
		}
		runes := []rune(cleanTitle)
		if len(runes) > 60 {
			cleanTitle = string(runes[:60])
		}
		sess.DraftTask.Title = cleanTitle
		sess.State = models.StateCreatingTaskDesc
		h.fsm.Set(userID, sess)

		prompt := l.Add.FormDescPrompt
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), prompt).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard("", l)))
		return true

	case models.StateCreatingTaskDesc:
		cleanDesc := strings.TrimSpace(emoji.StripCustomEmojis(text))
		if cleanDesc != "-" && cleanDesc != "" {
			runes := []rune(cleanDesc)
			if len(runes) > 1500 {
				cleanDesc = string(runes[:1500])
			}
			sess.DraftTask.Description = cleanDesc
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
		cleanTitle := strings.Join(strings.Fields(strings.ReplaceAll(text, "\n", " ")), " ")
		cleanTitle = emoji.StripCustomEmojis(cleanTitle)
		if cleanTitle == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		runes := []rune(cleanTitle)
		if len(runes) > 60 {
			cleanTitle = string(runes[:60])
		}
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Title = cleanTitle
			_ = h.storage.UpdateTask(ctx, task)
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TitleUpdated, task.ID, html.EscapeString(cleanTitle))).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, task.ID), fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingDesc:
		cleanDesc := strings.TrimSpace(emoji.StripCustomEmojis(text))
		if cleanDesc == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		runes := []rune(cleanDesc)
		if len(runes) > 1500 {
			cleanDesc = string(runes[:1500])
		}
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Description = cleanDesc
			_ = h.storage.UpdateTask(ctx, task)
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.DescUpdated, task.ID)).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, task.ID), fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingLabels:
		cleanText := strings.TrimSpace(text)
		if cleanText == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			if cleanText == "-" {
				task.Labels = nil
			} else {
				var rawParts []string
				if strings.Contains(cleanText, ",") {
					rawParts = strings.Split(cleanText, ",")
				} else {
					rawParts = strings.Fields(cleanText)
				}
				seen := make(map[string]bool)
				var labels []string
				for _, part := range rawParts {
					cleaned := strings.TrimSpace(part)
					cleaned = strings.TrimPrefix(cleaned, "#")
					cleaned = strings.ToLower(strings.TrimSpace(cleaned))
					cleaned = emoji.StripCustomEmojis(cleaned)
					if cleaned != "" && !seen[cleaned] {
						seen[cleaned] = true
						labels = append(labels, cleaned)
					}
				}
				task.Labels = labels
			}
			_ = h.storage.UpdateTask(ctx, task)

			formatted := task.FormattedLabels()
			if formatted == "" {
				formatted = "<i>(нет)</i>"
			}
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.LabelsUpdated, task.ID, formatted)).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, task.ID), fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAddingSubtask:
		cleanText := strings.ReplaceAll(text, "\r", " ")
		cleanText = strings.ReplaceAll(cleanText, "\n", " ")
		cleanText = strings.ReplaceAll(cleanText, "\t", " ")
		cleanTitle := strings.Join(strings.Fields(cleanText), " ")
		cleanTitle = emoji.StripCustomEmojis(cleanTitle)
		if cleanTitle == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		runes := []rune(cleanTitle)
		if len(runes) > 150 {
			cleanTitle = string(runes[:150])
		}
		_, err := h.storage.AddSubtask(ctx, sess.TaskID, cleanTitle)
		if err == nil {
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.SubtaskAdded, sess.TaskID, html.EscapeString(cleanTitle))).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, sess.TaskID), fmt.Sprintf("view:%s", sess.TaskID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingSubtask:
		cleanText := strings.ReplaceAll(text, "\r", " ")
		cleanText = strings.ReplaceAll(cleanText, "\n", " ")
		cleanText = strings.ReplaceAll(cleanText, "\t", " ")
		cleanTitle := strings.Join(strings.Fields(cleanText), " ")
		cleanTitle = emoji.StripCustomEmojis(cleanTitle)
		if cleanTitle == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		runes := []rune(cleanTitle)
		if len(runes) > 150 {
			cleanTitle = string(runes[:150])
		}
		err := h.storage.UpdateSubtask(ctx, sess.SubtaskID, cleanTitle)
		if err == nil {
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.SubtaskUpdated, sess.TaskID, html.EscapeString(cleanTitle))).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, sess.TaskID), fmt.Sprintf("view:%s", sess.TaskID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAddingComment:
		cleanComment := strings.TrimSpace(emoji.StripCustomEmojis(text))
		if cleanComment == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		runes := []rune(cleanComment)
		if len(runes) > 2000 {
			cleanComment = string(runes[:2000])
		}
		authorName := msg.From.Username
		if authorName == "" {
			authorName = msg.From.FirstName
		}
		_, err := h.storage.AddComment(ctx, sess.TaskID, userID, authorName, cleanComment)
		if err == nil {
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.CommentAdded, sess.TaskID)).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, sess.TaskID), fmt.Sprintf("view:%s", sess.TaskID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingComment:
		cleanComment := strings.TrimSpace(emoji.StripCustomEmojis(text))
		if cleanComment == "" {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}
		runes := []rune(cleanComment)
		if len(runes) > 2000 {
			cleanComment = string(runes[:2000])
		}
		err := h.storage.UpdateComment(ctx, sess.CommentID, cleanComment)
		if err == nil {
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.CommentUpdated, sess.TaskID)).
				WithParseMode(telego.ModeHTML).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, sess.TaskID), fmt.Sprintf("view:%s", sess.TaskID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
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

		if targetUID == 0 && text != "" {
			cleanText := strings.TrimPrefix(text, "@")
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
			if text == "" {
				_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Common.OnlyTextAllowed).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
				return true
			}
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TransferUserNotFound, text)).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		}

		senderName := msg.From.Username
		if senderName == "" {
			senderName = msg.From.FirstName
		}

		// Send offer DM to target user
		offerText := fmt.Sprintf(l.Edit.TransferOfferReceived, senderName, task.ID, html.EscapeString(task.Title), html.EscapeString(task.Description))
		inviteKb := BuildTransferInviteKeyboard(task.ID, targetUID, l)
		_, err = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(targetUID), offerText).WithParseMode(telego.ModeHTML).WithReplyMarkup(inviteKb))
		if err != nil {
			_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TransferUserNotFound, targetUsername)).WithReplyMarkup(BuildCancelKeyboard(sess.TaskID, l)))
			return true
		} else {
			msgReply := tu.Message(tu.ID(userID), fmt.Sprintf(l.Edit.TransferOfferSent, targetUsername)).
				WithReplyMarkup(sanitizeKeyboard(&telego.InlineKeyboardMarkup{
					InlineKeyboard: [][]telego.InlineKeyboardButton{
						{emoji.MakeInlineButton(fmt.Sprintf(l.Buttons.BackToTask, task.ID), fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", "")},
					},
				}))
			_, _ = SendMessageSafe(ctx, h.bot, msgReply)
		}
		h.fsm.Clear(userID)
		return true
	}

	return false
}


func (h *EditHandler) FSM() *FSM {
	return h.fsm
}

func (h *EditHandler) getAvailableTags(ctx context.Context, task *models.Task) []string {
	tagMap := make(map[string]bool)
	var ordered []string

	add := func(raw string) {
		clean := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "#")))
		if clean != "" && !tagMap[clean] {
			tagMap[clean] = true
			ordered = append(ordered, clean)
		}
	}

	for _, t := range h.cfg.Labels {
		add(t)
	}
	if storageTags, err := h.storage.GetAllLabels(ctx); err == nil {
		for _, t := range storageTags {
			add(t)
		}
	}
	if task != nil {
		for _, t := range task.Labels {
			add(t)
		}
	}
	return ordered
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

