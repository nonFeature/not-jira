package bot

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"not-jira/internal/config"
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
	isAdmin := h.cfg.Telegram.IsAdmin(userID)

	if strings.HasPrefix(data, "set_status:") {
		// set_status:{task_id}:{status}
		parts := strings.Split(data, ":")
		if len(parts) != 3 {
			return
		}
		if !isAdmin {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("⛔️ Менять статус могут только администраторы").WithShowAlert())
			return
		}

		taskID := parts[1]
		newStatus := models.TaskStatus(parts[2])

		task, err := h.storage.GetTask(ctx, taskID)
		if err != nil || task == nil {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Задача не найдена"))
			return
		}

		oldStatus := task.Status
		task.Status = newStatus
		if err := h.storage.UpdateTask(ctx, task); err != nil {
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Ошибка обновления статуса"))
			return
		}

		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText(fmt.Sprintf("Статус изменен: %s %s", newStatus.Emoji(), newStatus.Russian())))

		// Notify topic & author if status changed
		if oldStatus != newStatus {
			h.notifier.NotifyStatusChange(ctx, task)
		}

		// Update the current message card
		h.updateMessageCard(query.Message.GetChat().ID, query.Message.GetMessageID(), task, isAdmin)
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
			_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("❌ Ошибка переключения"))
			return
		}

		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))

		task, _ := h.storage.GetTask(ctx, taskID)
		if task != nil {
			h.updateMessageCard(query.Message.GetChat().ID, query.Message.GetMessageID(), task, isAdmin)
		}
		return
	}

	// GitHub Issues style edit actions
	if !isAdmin {
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("⛔️ Только для администраторов").WithShowAlert())
		return
	}

	if strings.HasPrefix(data, "edit_title:") {
		taskID := strings.TrimPrefix(data, "edit_title:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateEditingTitle,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("✏️ Введите новый <b>заголовок</b> для задачи <code>[%s]</code>:", taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "edit_desc:") {
		taskID := strings.TrimPrefix(data, "edit_desc:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateEditingDesc,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("📝 Введите новое <b>описание</b> для задачи <code>[%s]</code>:", taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "add_sub:") {
		taskID := strings.TrimPrefix(data, "add_sub:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAddingSubtask,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("➕ Введите название <b>подзадачи</b> для <code>[%s]</code>:", taskID)).WithParseMode(telego.ModeHTML))
		return
	}

	if strings.HasPrefix(data, "add_comm:") {
		taskID := strings.TrimPrefix(data, "add_comm:")
		h.fsm.Set(userID, &models.UserSession{
			State:  models.StateAddingComment,
			TaskID: taskID,
		})
		_ = h.bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID))
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("💬 Введите текст <b>комментария</b> для задачи <code>[%s]</code>:", taskID)).WithParseMode(telego.ModeHTML))
		return
	}
}

func (h *EditHandler) HandleFSMMessage(ctx context.Context, msg *telego.Message) bool {
	userID := msg.From.ID
	sess := h.fsm.Get(userID)
	if sess == nil || sess.State == models.StateNone {
		return false
	}

	text := strings.TrimSpace(msg.Text)
	if text == "/cancel" {
		h.fsm.Clear(userID)
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), "Действие отменено."))
		return true
	}

	switch sess.State {
	case models.StateCreatingTaskTitle:
		sess.DraftTask.Title = text
		sess.State = models.StateCreatingTaskDesc
		h.fsm.Set(userID, sess)

		prompt := "✅ Заголовок сохранен.\n\nТеперь введите <b>описание</b> задачи (или отправьте <code>-</code>, чтобы оставить исходный текст):"
		_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), prompt).WithParseMode(telego.ModeHTML))
		return true

	case models.StateCreatingTaskDesc:
		if text != "-" {
			sess.DraftTask.Description = text
		}

		if err := h.storage.CreateTask(ctx, sess.DraftTask); err != nil {
			_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("❌ Ошибка сохранения в БД: %v", err)))
			h.fsm.Clear(userID)
			return true
		}

		task := sess.DraftTask
		h.fsm.Clear(userID)

		cardHTML := RenderTaskCard(task)
		kb := BuildTaskInlineKeyboard(task, true)

		// 1. If origin chat was a group/topic, send brief confirmation to user in topic
		if task.ChatID != userID {
			itemWord := "баг"
			if task.Type == models.TaskTypeIdea {
				itemWord = "идея"
			}
			topicReplyText := fmt.Sprintf("✅ Ваш(а) %s <b>[%s]</b> принят(а) в обработку!", itemWord, task.ID)
			topicMsg := tu.Message(tu.ID(task.ChatID), topicReplyText).WithParseMode(telego.ModeHTML)
			if task.MessageID != 0 {
				topicMsg.ReplyParameters = &telego.ReplyParameters{MessageID: int(task.MessageID)}
			}
			if task.TopicID != 0 {
				topicMsg.MessageThreadID = int(task.TopicID)
			}
			_, _ = h.bot.SendMessage(topicMsg)
		}

		// 2. Send management card to admin in DM
		confirmDM := tu.Message(tu.ID(userID), "🎉 <b>Задача успешно создана!</b>\n\n"+cardHTML).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(kb)
		_, _ = h.bot.SendMessage(confirmDM)
		return true

	case models.StateEditingTitle:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Title = text
			_ = h.storage.UpdateTask(ctx, task)
			_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("✅ Заголовок задачи <b>[%s]</b> обновлен на: %s", task.ID, html.EscapeString(text))).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateEditingDesc:
		task, err := h.storage.GetTask(ctx, sess.TaskID)
		if err == nil && task != nil {
			task.Description = text
			_ = h.storage.UpdateTask(ctx, task)
			_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("✅ Описание задачи <b>[%s]</b> обновлено.", task.ID)).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true

	case models.StateAddingSubtask:
		_, err := h.storage.AddSubtask(ctx, sess.TaskID, text)
		if err == nil {
			_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("✅ Подзадача добавлена к <b>[%s]</b>: %s", sess.TaskID, html.EscapeString(text))).WithParseMode(telego.ModeHTML))
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
			_, _ = h.bot.SendMessage(tu.Message(tu.ID(userID), fmt.Sprintf("✅ Комментарий добавлен к <b>[%s]</b>.", sess.TaskID)).WithParseMode(telego.ModeHTML))
		}
		h.fsm.Clear(userID)
		return true
	}

	return false
}

func (h *EditHandler) updateMessageCard(chatID int64, messageID int, task *models.Task, isAdmin bool) {
	cardHTML := RenderTaskCard(task)
	kb := BuildTaskInlineKeyboard(task, isAdmin)

	editParams := &telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		Text:        cardHTML,
		ParseMode:   telego.ModeHTML,
		ReplyMarkup: kb,
	}
	_, _ = h.bot.EditMessageText(editParams)
}
