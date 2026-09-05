package locales

import (
	"fmt"
	"not-jira/internal/emoji"
)

func GetRu() *Bundle {
	return &Bundle{
		Common: Common{
			AdminOnly: fmt.Sprintf("%s Доступно только администраторам.", emoji.Lock()),
			Cancelled: "Отменено.",
		},
		Start: Start{
			GreetingUser: fmt.Sprintf("%s <b>not-jira</b> — трекер задач в топиках\n\n"+
				"<b>Команды:</b>\n"+
				"/list — список задач\n"+
				"/view [ID] — карточка задачи (напр. /view B0)\n"+
				"/settings — уведомления в ЛС\n\n", emoji.Wave()),
			GreetingAdmin: fmt.Sprintf("%s <b>Админам:</b>\n"+
				"/add — создать задачу (ответом на сообщение)\n", emoji.Star()),
		},
		Settings: Settings{
			Title:          fmt.Sprintf("%s <b>Настройки уведомлений</b>\n\nУведомления в ЛС: <b>%%s</b>", emoji.Gear()),
			StatusEnabled:  fmt.Sprintf("включены %s", emoji.Bell()),
			StatusDisabled: fmt.Sprintf("отключены %s", emoji.BellOff()),
			BtnEnable:      "Включить",
			BtnDisable:     "Отключить",
			UpdatedAlert:   "Настройки обновлены",
		},
		Add: Add{
			UsageHint:          fmt.Sprintf("%s Ответьте /add на сообщение или укажите текст: /add [описание]", emoji.Info()),
			ProcessingAI:       fmt.Sprintf("%s <i>Обработка...</i>", emoji.Hourglass()),
			AcceptedBug:        fmt.Sprintf("%s Баг <b>[%%s]</b> принят", emoji.Check()),
			AcceptedIdea:       fmt.Sprintf("%s Идея <b>[%%s]</b> принята", emoji.Check()),
			CardHeader:         fmt.Sprintf("%s <b>Карточка задачи:</b>\n\n", emoji.Clipboard()),
			FormTitlePrompt:    fmt.Sprintf("%s <b>Новая задача <code>[%%s]</code> (%%s)</b>\n\n<blockquote>%%s</blockquote>\n\nВведите <b>заголовок</b> (до 60 символов):", emoji.Memo()),
			FormDescPrompt:     fmt.Sprintf("%s Заголовок сохранен.\n\nВведите <b>описание</b> (или <code>-</code> для исходного текста):", emoji.Check()),
			FormCreatedSuccess: fmt.Sprintf("%s <b>Задача создана!</b>\n\n", emoji.Party()),
		},
		Edit: Edit{
			PromptEditTitle:        fmt.Sprintf("%s Новый <b>заголовок</b> для <code>[%%s]</code>:", emoji.Pencil()),
			PromptEditDesc:         fmt.Sprintf("%s Новое <b>описание</b> для <code>[%%s]</code>:", emoji.Memo()),
			PromptAddSubtask:       fmt.Sprintf("%s Новая <b>подзадача</b> для <code>[%%s]</code>:", emoji.Plus()),
			PromptAddComment:       fmt.Sprintf("%s Новый <b>комментарий</b> для <code>[%%s]</code>:", emoji.Messages()),
			TitleUpdated:           fmt.Sprintf("%s Заголовок <b>[%%s]</b> обновлен: %%s", emoji.Check()),
			DescUpdated:            fmt.Sprintf("%s Описание <b>[%%s]</b> обновлено.", emoji.Check()),
			SubtaskAdded:           fmt.Sprintf("%s Подзадача добавлена в <b>[%%s]</b>: %%s", emoji.Check()),
			CommentAdded:           fmt.Sprintf("%s Комментарий добавлен в <b>[%%s]</b>.", emoji.Check()),
			StatusChangedAlert:     "Статус: %s %s",
			StatusReopenedAlert:    "Задача открыта заново",
			TaskArchivedAlert:      "Задача перемещена в архив",
			PromptTransfer:         fmt.Sprintf("%s Кому передать задачу <code>[%%s]</code>?\nВведите @username или перешлите сообщение пользователя:", emoji.User()),
			TransferUserNotFound:   "Пользователь @%s не найден. Попросите его написать боту /start в ЛС.",
			TransferOfferSent:      "Предложение передать задачу отправлено @%s в ЛС.",
			TransferOfferReceived:  fmt.Sprintf("%s @%%s предлагает вам задачу <b>[%%s]</b>:\n\n<b>%%s</b>\n<blockquote>%%s</blockquote>\n\nВы хотите взять задачу?", emoji.User()),
			TransferAcceptedNotify: fmt.Sprintf("%s @%%s принял задачу <b>[%%s]</b>.", emoji.Check()),
			TransferRejectedNotify: fmt.Sprintf("%s @%%s отклонил предложение по задаче <b>[%%s]</b>.", emoji.Cross()),
			TaskClaimedNotify:      fmt.Sprintf("%s @%%s взял задачу <b>[%%s]</b> в работу.", emoji.Check()),
		},
		View: View{
			UsageHint:  fmt.Sprintf("%s Пример: /view B0", emoji.Info()),
			NotFound:   fmt.Sprintf("%s Задача <code>%%s</code> не найдена.", emoji.Cross()),
			ListHeader: fmt.Sprintf("%s <b>Задачи (%%d)</b>\nФильтр: <code>%%s</code> | <code>%%s</code>\n\n", emoji.Clipboard()),
			PageFormat: "Стр. %d/%d",
		},
		Task: Task{
			Header:          "<b>[%s]</b> %s <b>%s: %s</b>\n\n",
			StatusLabel:     "<b>Статус:</b> %s %s\n",
			AuthorLabel:     "<b>Автор:</b> @%s\n",
			AssigneeLabel:   "<b>Исполнитель:</b> @%s\n",
			UnassignedLabel: "<b>Исполнитель:</b> <i>не назначен</i>\n",
			ArchivedBadge:   fmt.Sprintf(" %s <i>(в архиве)</i>", emoji.Box()),
			CreatedLabel:    "<b>Создано:</b> %s UTC\n\n",
			DescLabel:       "<b>Описание:</b>\n<blockquote expandable>%s</blockquote>\n\n",
			SubtasksLabel:   "<b>Подзадачи (%d/%d):</b>\n",
			CommentsLabel:   "<b>Комментарии (%d):</b>\n",
			TypeBug:         "Баг",
			TypeIdea:        "Идея",
			StatusNew:       "Новое",
			StatusProgress:  "В работе",
			StatusDone:      "Готово",
			StatusRejected:  "Отклонено",
		},
		Buttons: Buttons{
			InProgress:   "В работу",
			Done:         "Готово",
			Rejected:     "Отклонить",
			Reopen:       "Открыть заново",
			Archive:      "В архив",
			Claim:        "Взять себе",
			Transfer:     "Передать",
			AcceptTask:   "Взять задачу",
			RejectTask:   "Отклонить",
			EditTitle:    "Заголовок",
			EditDesc:     "Описание",
			AddSubtask:   "Саб-таск",
			AddComment:   "Комментарий",
			OriginalPost: "Исходный пост",
			Refresh:      "Обновить",
			PrevPage:     "Назад",
			NextPage:     "Вперед",
			OpenDM:       "Открыть ЛС",
		},
		Filters: Filters{
			AllTypes:    "Все",
			Bugs:        "Баги",
			Ideas:       "Идеи",
			AllStatuses: "Все статусы",
			New:         "Новые",
			InProgress:  "В работе",
			Done:        "Готово",
		},
		Notifications: Notifications{
			PromptStartDM:      fmt.Sprintf("%s %%s, начните диалог с ботом в ЛС.", emoji.Warning()),
			TopicStatusUpdated: fmt.Sprintf("%s <b>[%%s]</b> %%s %%s\n%%s", emoji.Bell()),
			DMStatusUpdated:    fmt.Sprintf("%s <b>[%%s]</b> %%s %%s\n%%s", emoji.Bell()),
		},
	}
}
