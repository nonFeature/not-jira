package bot

import (
	"context"
	"fmt"
	"html"
	"strings"

	"not-jira/internal/ai"
	"not-jira/internal/config"
	"not-jira/internal/locales"
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
	l := locales.ForUser(msg.From.LanguageCode)
	senderID := msg.From.ID
	if !h.cfg.Telegram.IsAdmin(senderID) {
		reply := tu.Message(tu.ID(senderID), l.Common.AdminOnly)
		_, _ = SendMessageSafe(ctx, h.bot, reply)
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
		reply := tu.Message(tu.ID(senderID), l.Add.UsageHint).WithParseMode(telego.ModeHTML)
		_, _ = SendMessageSafe(ctx, h.bot, reply)
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
		if strings.Contains(lower, "идея") || strings.Contains(lower, "предлагаю") || strings.Contains(lower, "фича") || strings.Contains(lower, "idea") || strings.Contains(lower, "feature") {
			taskType = models.TaskTypeIdea
		}
	}

	// Calculate Next Task ID (e.g. B0, I0)
	nextID, nextNum, err := h.storage.GetNextTaskID(ctx, taskType)
	if err != nil {
		reply := tu.Message(tu.ID(senderID), fmt.Sprintf("❌ ID generation error: %v", err))
		_, _ = SendMessageSafe(ctx, h.bot, reply)
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
		statusMsg, _ := SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(senderID), l.Add.ProcessingAI).WithParseMode(telego.ModeHTML))

		typeLabel := TaskTypeName(taskType, l)
		res, err := h.summarizer.Summarize(ctx, typeLabel, sourceText)
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
			_ = h.bot.DeleteMessage(ctx, &telego.DeleteMessageParams{
				ChatID:    tu.ID(senderID),
				MessageID: statusMsg.MessageID,
			})
		}

		if err := h.storage.CreateTask(ctx, draft); err != nil {
			reply := tu.Message(tu.ID(senderID), fmt.Sprintf("❌ Database save error: %v", err))
			_, _ = SendMessageSafe(ctx, h.bot, reply)
			return
		}

		cardHTML := RenderTaskCard(draft, l)
		kb := BuildTaskInlineKeyboard(draft, senderID, true, h.cfg.Telegram.IsDev(senderID), l)

		// 1. In group/topic: reply briefly that the task was accepted
		if msg.Chat.ID != senderID {
			var topicReplyText string
			if draft.Type == models.TaskTypeIdea {
				topicReplyText = fmt.Sprintf(l.Add.AcceptedIdea, draft.ID)
			} else {
				topicReplyText = fmt.Sprintf(l.Add.AcceptedBug, draft.ID)
			}
			topicMsg := tu.Message(tu.ID(msg.Chat.ID), topicReplyText).WithParseMode(telego.ModeHTML)
			if origMsgID != 0 {
				topicMsg.ReplyParameters = &telego.ReplyParameters{MessageID: origMsgID}
			}
			if topicID != 0 {
				topicMsg.MessageThreadID = int(topicID)
			}
			_, _ = SendMessageSafe(ctx, h.bot, topicMsg)
		}

		// 2. In admin's DM: send the full management card with buttons
		dmCardMsg := tu.Message(tu.ID(senderID), l.Add.CardHeader+cardHTML).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(kb)
		_, dmErr := SendMessageSafe(ctx, h.bot, dmCardMsg)
		if dmErr != nil && msg.Chat.ID != senderID {
			PromptStartInDM(ctx, h.bot, h.botUsername, msg)
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

	dmText := fmt.Sprintf(l.Add.FormTitlePrompt,
		draft.ID, TaskTypeName(draft.Type, l), html.EscapeString(sourceText))

	dmMsg := tu.Message(tu.ID(senderID), dmText).WithParseMode(telego.ModeHTML).WithReplyMarkup(BuildCancelKeyboard("", l))
	_, dmErr := SendMessageSafe(ctx, h.bot, dmMsg)

	if dmErr != nil {
		// Admin hasn't started bot in DM yet
		if msg.Chat.ID != senderID {
			PromptStartInDM(ctx, h.bot, h.botUsername, msg)
		}
		h.fsm.Clear(senderID)
		return
	}
}
