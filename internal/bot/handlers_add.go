package bot

import (
	"context"
	"fmt"
	"html"
	"strings"

	"not-jira/internal/ai"
	"not-jira/internal/config"
	"not-jira/internal/models"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type AddHandler struct {
	bot         *telego.Bot
	botUsername string
	cfg         *config.Config
	storage     storage.Storage
	summarizer  *ai.Summarizer
	fsm         *FSM
}

func NewAddHandler(bot *telego.Bot, botUsername string, cfg *config.Config, storage storage.Storage, summarizer *ai.Summarizer, fsm *FSM) *AddHandler {
	return &AddHandler{
		bot:         bot,
		botUsername: botUsername,
		cfg:         cfg,
		storage:     storage,
		summarizer:  summarizer,
		fsm:         fsm,
	}
}

func (h *AddHandler) Handle(ctx context.Context, msg *telego.Message) {
	senderID := msg.From.ID
	if !h.cfg.Telegram.IsAdmin(senderID) {
		reply := tu.Message(tu.ID(senderID), "⛔️ Создавать задачи могут только администраторы.")
		_, _ = h.bot.SendMessage(reply)
		return
	}

	var sourceText string
	var authorID int64
	var authorUsername string
	var origMsgID int
	var topicID int64 = int64(msg.MessageThreadID)

	if msg.ReplyToMessage != nil {
		replyMsg := msg.ReplyToMessage
		sourceText = replyMsg.Text
		if sourceText == "" {
			sourceText = replyMsg.Caption
		}
		if replyMsg.From != nil {
			authorID = replyMsg.From.ID
			authorUsername = replyMsg.From.Username
			if authorUsername == "" {
				authorUsername = strings.TrimSpace(replyMsg.From.FirstName + " " + replyMsg.From.LastName)
			}
		}
		origMsgID = replyMsg.MessageID
	} else {
		// Take text after /add
		parts := strings.SplitN(msg.Text, " ", 2)
		if len(parts) > 1 {
			sourceText = strings.TrimSpace(parts[1])
		}
		authorID = msg.From.ID
		authorUsername = msg.From.Username
		origMsgID = msg.MessageID
	}

	if sourceText == "" {
		reply := tu.Message(tu.ID(senderID), "ℹ️ Ответьте командой <code>/add</code> на сообщение с багом/идеей, либо укажите текст: <code>/add Описание бага</code>").WithParseMode(telego.ModeHTML)
		_, _ = h.bot.SendMessage(reply)
		return
	}

	// Detect task type based on topic ID
	taskType := models.TaskTypeBug
	if msg.MessageThreadID != 0 && msg.MessageThreadID == h.cfg.Telegram.IdeasTopicID {
		taskType = models.TaskTypeIdea
	} else if msg.MessageThreadID != 0 && msg.MessageThreadID == h.cfg.Telegram.BugsTopicID {
		taskType = models.TaskTypeBug
	} else {
		// If in private or generic chat, check if contains idea keywords
		lower := strings.ToLower(sourceText)
		if strings.Contains(lower, "идея") || strings.Contains(lower, "предлагаю") || strings.Contains(lower, "фича") {
			taskType = models.TaskTypeIdea
		}
	}

	// Calculate Next Task ID (e.g. B0, I0)
	nextID, nextNum, err := h.storage.GetNextTaskID(ctx, taskType)
	if err != nil {
		reply := tu.Message(tu.ID(senderID), fmt.Sprintf("❌ Ошибка генерации ID задачи: %v", err))
		_, _ = h.bot.SendMessage(reply)
		return
	}

	// Generate message link
	msgLink := ""
	if msg.Chat.Username != "" {
		msgLink = fmt.Sprintf("https://t.me/%s/%d", msg.Chat.Username, origMsgID)
	} else if msg.Chat.ID < 0 {
		cleanID := fmt.Sprintf("%d", msg.Chat.ID)
		cleanID = strings.TrimPrefix(cleanID, "-100")
		cleanID = strings.TrimPrefix(cleanID, "-")
		if topicID != 0 {
			msgLink = fmt.Sprintf("https://t.me/c/%s/%d/%d", cleanID, topicID, origMsgID)
		} else {
			msgLink = fmt.Sprintf("https://t.me/c/%s/%d", cleanID, origMsgID)
		}
	}

	draft := &models.Task{
		ID:             nextID,
		Num:            nextNum,
		Type:           taskType,
		Status:         models.StatusNew,
		ChatID:         msg.Chat.ID,
		TopicID:        topicID,
		MessageID:      int64(origMsgID),
		MessageLink:    msgLink,
		AuthorID:       authorID,
		AuthorUsername: authorUsername,
	}

	// If AI is enabled: auto-summarize
	if h.summarizer != nil && h.cfg.AI.Enabled {
		statusMsg, _ := h.bot.SendMessage(tu.Message(tu.ID(senderID), "⏳ <i>Обрабатываю задачу через нейросеть...</i>").WithParseMode(telego.ModeHTML))

		res, err := h.summarizer.Summarize(ctx, taskType.Russian(), sourceText)
		if err == nil && res != nil {
			draft.Title = res.Title
			draft.Description = res.Description
		} else {
			// Fallback if AI fails
			lines := strings.SplitN(sourceText, "\n", 2)
			draft.Title = strings.TrimSpace(lines[0])
			draft.Description = sourceText
		}

		if statusMsg != nil {
			_ = h.bot.DeleteMessage(&telego.DeleteMessageParams{
				ChatID:    tu.ID(senderID),
				MessageID: statusMsg.MessageID,
			})
		}

		if err := h.storage.CreateTask(ctx, draft); err != nil {
			reply := tu.Message(tu.ID(senderID), fmt.Sprintf("❌ Ошибка сохранения в БД: %v", err))
			_, _ = h.bot.SendMessage(reply)
			return
		}

		cardHTML := RenderTaskCard(draft)
		kb := BuildTaskInlineKeyboard(draft, true)

		// 1. In group/topic: reply briefly that the task was accepted
		if msg.Chat.ID != senderID {
			itemWord := "баг"
			if draft.Type == models.TaskTypeIdea {
				itemWord = "идея"
			}
			topicReplyText := fmt.Sprintf("✅ Ваш(а) %s <b>[%s]</b> принят(а) в обработку!", itemWord, draft.ID)
			topicMsg := tu.Message(tu.ID(msg.Chat.ID), topicReplyText).WithParseMode(telego.ModeHTML)
			if origMsgID != 0 {
				topicMsg.ReplyParameters = &telego.ReplyParameters{MessageID: origMsgID}
			}
			if topicID != 0 {
				topicMsg.MessageThreadID = int(topicID)
			}
			_, _ = h.bot.SendMessage(topicMsg)
		}

		// 2. In admin's DM: send the full management card with buttons
		dmCardMsg := tu.Message(tu.ID(senderID), "📋 <b>Управление задачей:</b>\n\n"+cardHTML).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(kb)
		_, dmErr := h.bot.SendMessage(dmCardMsg)
		if dmErr != nil && msg.Chat.ID != senderID {
			PromptStartInDM(h.bot, h.botUsername, msg)
		}
		return
	}

	// If AI is disabled: Launch interactive form in DM with admin
	draft.Description = sourceText // save original text as default description
	h.fsm.Set(senderID, &models.UserSession{
		State:     models.StateCreatingTaskTitle,
		TaskID:    draft.ID,
		DraftTask: draft,
	})

	dmText := fmt.Sprintf("📝 <b>Создание задачи <code>[%s]</code> (%s)</b>\n\n<b>Исходный текст:</b>\n<blockquote>%s</blockquote>\n\nОтправьте в ответ <b>заголовок</b> задачи (до 60 символов):",
		draft.ID, draft.Type.Russian(), html.EscapeString(sourceText))

	dmMsg := tu.Message(tu.ID(senderID), dmText).WithParseMode(telego.ModeHTML)
	_, dmErr := h.bot.SendMessage(dmMsg)

	if dmErr != nil {
		// Admin hasn't started bot in DM yet
		if msg.Chat.ID != senderID {
			PromptStartInDM(h.bot, h.botUsername, msg)
		}
		h.fsm.Clear(senderID)
		return
	}
}
