package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"not-jira/internal/config"
	"not-jira/internal/locales"
	"not-jira/internal/models"
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

	startMsg := tu.Message(tu.ID(msg.From.ID), text).WithParseMode(telego.ModeHTML)
	startMsg.LinkPreviewOptions = &telego.LinkPreviewOptions{IsDisabled: true}
	_, err := SendMessageSafe(ctx, h.bot, startMsg)
	if err != nil && msg.Chat.ID != msg.From.ID {
		PromptStartInDM(ctx, h.bot, h.botUsername, msg)
	}
}

func (h *UserHandler) formatSettings(settings *models.UserSettings, l *locales.Bundle) (string, *telego.InlineKeyboardMarkup) {
	dmDesc := l.Settings.StatusEnabled
	if !settings.NotifyDM {
		dmDesc = l.Settings.StatusDisabled
	}

	forumDesc := l.Settings.StatusSound
	if !settings.NotifyForum {
		forumDesc = l.Settings.StatusSilent
	}

	text := fmt.Sprintf(l.Settings.Title, dmDesc, forumDesc)
	kb := BuildSettingsKeyboard(settings.NotifyDM, settings.NotifyForum, l)
	return text, kb
}

func (h *UserHandler) HandleSettings(ctx context.Context, msg *telego.Message) {
	l := locales.ForUser(msg.From.LanguageCode)
	settings, err := h.storage.GetUserSettings(ctx, msg.From.ID)
	if err != nil {
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(msg.From.ID), "❌ Error loading settings."))
		return
	}

	text, kb := h.formatSettings(settings, l)
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

	if query.Data == "settings:toggle_forum" {
		newStatus := !settings.NotifyForum
		if err := h.storage.SetNotifyForum(ctx, userID, newStatus); err != nil {
			_ = h.bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("❌ Save error"))
			return
		}
		settings.NotifyForum = newStatus
	} else {
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
		settings.NotifyDM = newStatus
	}

	text, kb := h.formatSettings(settings, l)
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

func (h *UserHandler) HandleBackup(ctx context.Context, msg *telego.Message) {
	l := locales.ForUser(msg.From.LanguageCode)
	userID := msg.From.ID

	if !h.cfg.Telegram.IsAdmin(userID) {
		reply := tu.Message(tu.ID(userID), l.Common.AdminOnly).WithParseMode(telego.ModeHTML)
		_, _ = SendMessageSafe(ctx, h.bot, reply)
		return
	}

	statusMsg, _ := SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Backup.Creating))

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	tempFileName := fmt.Sprintf("not-jira-backup-%s.db", timestamp)
	tempPath := filepath.Join(os.TempDir(), tempFileName)

	if err := h.storage.Backup(ctx, tempPath); err != nil {
		log.Printf("[Backup ERROR] Backup failed: %v", err)
		if statusMsg != nil {
			_ = h.bot.DeleteMessage(ctx, &telego.DeleteMessageParams{
				ChatID:    tu.ID(userID),
				MessageID: statusMsg.MessageID,
			})
		}
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Backup.Failed))
		return
	}
	defer os.Remove(tempPath)

	fileInfo, err := os.Stat(tempPath)
	if err != nil {
		log.Printf("[Backup ERROR] Stat backup failed: %v", err)
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Backup.Failed))
		return
	}

	_, totalTasks, _ := h.storage.ListTasks(ctx, storage.TaskFilter{IncludeArchived: true}, 0, 1)

	fileSizeStr := formatFileSize(fileInfo.Size())
	caption := fmt.Sprintf(l.Backup.Caption, time.Now().Format("02.01.2006 15:04:05"), totalTasks, fileSizeStr)

	f, err := os.Open(tempPath)
	if err != nil {
		log.Printf("[Backup ERROR] Open backup failed: %v", err)
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Backup.Failed))
		return
	}
	defer f.Close()

	doc := tu.Document(tu.ID(userID), tu.File(f)).WithCaption(caption).WithParseMode(telego.ModeHTML)
	_, err = h.bot.SendDocument(ctx, doc)

	if statusMsg != nil {
		_ = h.bot.DeleteMessage(ctx, &telego.DeleteMessageParams{
			ChatID:    tu.ID(userID),
			MessageID: statusMsg.MessageID,
		})
	}

	if err != nil {
		log.Printf("[Backup ERROR] SendDocument failed: %v", err)
		_, _ = SendMessageSafe(ctx, h.bot, tu.Message(tu.ID(userID), l.Backup.Failed))
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
