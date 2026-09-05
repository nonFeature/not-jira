package bot

import (
	"context"
	"fmt"
	"html"

	"not-jira/internal/locales"
	"not-jira/internal/models"
	"not-jira/internal/storage"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Notifier struct {
	bot     *telego.Bot
	storage storage.Storage
}

func NewNotifier(bot *telego.Bot, storage storage.Storage) *Notifier {
	return &Notifier{
		bot:     bot,
		storage: storage,
	}
}

func (n *Notifier) NotifyStatusChange(ctx context.Context, task *models.Task) {
	// Only notify when status is IN_PROGRESS ("в работе"), DONE ("готово") or REJECTED ("отклонено")
	if task.Status != models.StatusInProgress && task.Status != models.StatusDone && task.Status != models.StatusRejected {
		return
	}

	// Topic notification (using Russian as community default)
	lTopic := locales.GetRu()
	statusText := fmt.Sprintf(lTopic.Notifications.TopicStatusUpdated,
		task.ID, TaskStatusEmoji(task.Status), TaskStatusName(task.Status, lTopic), html.EscapeString(task.Title))

	var authorSettings *models.UserSettings
	if task.AuthorID != 0 {
		settings, err := n.storage.GetUserSettings(ctx, task.AuthorID)
		if err == nil {
			authorSettings = settings
		}
	}

	// 1. Reply to the original message in the forum topic (if chat_id & message_id are set)
	if task.ChatID != 0 && task.MessageID != 0 {
		msg := tu.Message(tu.ID(task.ChatID), statusText)
		msg.ReplyParameters = &telego.ReplyParameters{
			MessageID: int(task.MessageID),
		}
		msg.ParseMode = telego.ModeHTML
		if task.TopicID != 0 {
			msg.MessageThreadID = int(task.TopicID)
		}
		// If author configured forum notifications to be silent, send with no sound!
		if authorSettings != nil && !authorSettings.NotifyForum {
			msg.DisableNotification = true
		}
		_, _ = SendMessageSafe(ctx, n.bot, msg)
	}

	// 2. Notify author in private chat if enabled
	if task.AuthorID != 0 && authorSettings != nil && authorSettings.NotifyDM {
		lAuthor := locales.GetRu()
		dmText := fmt.Sprintf(lAuthor.Notifications.DMStatusUpdated,
			task.ID, TaskStatusEmoji(task.Status), TaskStatusName(task.Status, lAuthor), html.EscapeString(task.Title))

		dmMsg := tu.Message(tu.ID(task.AuthorID), dmText)
		dmMsg.ParseMode = telego.ModeHTML
		dmMsg.ReplyMarkup = BuildSettingsKeyboard(authorSettings.NotifyDM, authorSettings.NotifyForum, lAuthor)

		_, _ = SendMessageSafe(ctx, n.bot, dmMsg)
	}
}
