package locales

import (
	"fmt"
	"not-jira/internal/emoji"
)

func GetRu() *Bundle {
	return &Bundle{
		Common: Common{
			AdminOnly: fmt.Sprintf("%s Создавать задачи могут только администраторы.", emoji.Lock()),
			Cancelled: "Действие отменено.",
		},
		Start: Start{
			GreetingUser: fmt.Sprintf("%s <b>Привет! Я not-jira.</b>\n\n"+
				"Я помогаю собирать и отслеживать задачи, баги и идеи прямо в топиках форума.\n\n"+
				"<b>Команды:</b>\n"+
				"• /list — Просмотр списка задач с фильтрами и страницами\n"+
				"• /view [ID] — Просмотр карточки задачи (например, /view B0)\n"+
				"• /settings — Настройки уведомлений в ЛС\n\n", emoji.Wave()),
			GreetingAdmin: fmt.Sprintf("%s <b>Команды администратора:</b>\n"+
				"• /add — Ответьте на сообщение в топике форума, чтобы зарегистрировать баг (B) или идею (I)\n"+
				"• В карточке задачи вам доступны кнопки изменения статуса, заголовка, описания, саб-тасков и комментариев.\n", emoji.Star()),
		},
		Settings: Settings{
			Title:          fmt.Sprintf("%s <b>Настройки уведомлений</b>\n\nУведомления в личные сообщения: <b>%%s</b>\n\nВы можете изменить настройку нажатием кнопки:", emoji.Gear()),
			StatusEnabled:  fmt.Sprintf("включены %s", emoji.Bell()),
			StatusDisabled: fmt.Sprintf("отключены %s", emoji.BellOff()),
			BtnEnable:      fmt.Sprintf("%s Включить уведы в ЛС", emoji.Bell()),
			BtnDisable:     fmt.Sprintf("%s Отключить уведы в ЛС", emoji.BellOff()),
			UpdatedAlert:   "Настройки обновлены",
		},
		Add: Add{
			UsageHint:          fmt.Sprintf("%s Ответьте командой /add на сообщение с багом/идеей, либо укажите текст: /add Описание бага", emoji.Info()),
			ProcessingAI:       fmt.Sprintf("%s <i>Обрабатываю задачу через нейросеть...</i>", emoji.Hourglass()),
			AcceptedBug:        fmt.Sprintf("%s Ваш баг <b>[%%s]</b> принят в обработку!", emoji.Check()),
			AcceptedIdea:       fmt.Sprintf("%s Ваша идея <b>[%%s]</b> принята в обработку!", emoji.Check()),
			CardHeader:         fmt.Sprintf("%s <b>Управление задачей:</b>\n\n", emoji.Clipboard()),
			FormTitlePrompt:    fmt.Sprintf("%s <b>Создание задачи <code>[%%s]</code> (%%s)</b>\n\n<b>Исходный текст:</b>\n<blockquote>%%s</blockquote>\n\nОтправьте в ответ <b>заголовок</b> задачи (до 60 символов):", emoji.Memo()),
			FormDescPrompt:     fmt.Sprintf("%s Заголовок сохранен.\n\nТеперь введите <b>описание</b> задачи (или отправьте <code>-</code>, чтобы оставить исходный текст):", emoji.Check()),
			FormCreatedSuccess: fmt.Sprintf("%s <b>Задача успешно создана!</b>\n\n", emoji.Party()),
		},
		Edit: Edit{
			PromptEditTitle:    fmt.Sprintf("%s Введите новый <b>заголовок</b> для задачи <code>[%%s]</code>:", emoji.Pencil()),
			PromptEditDesc:     fmt.Sprintf("%s Введите новое <b>описание</b> для задачи <code>[%%s]</code>:", emoji.Memo()),
			PromptAddSubtask:   fmt.Sprintf("%s Введите название <b>подзадачи</b> для <code>[%%s]</code>:", emoji.Plus()),
			PromptAddComment:   fmt.Sprintf("%s Введите текст <b>комментария</b> для задачи <code>[%%s]</code>:", emoji.Messages()),
			TitleUpdated:       fmt.Sprintf("%s Заголовок задачи <b>[%%s]</b> обновлен на: %%s", emoji.Check()),
			DescUpdated:        fmt.Sprintf("%s Описание задачи <b>[%%s]</b> обновлено.", emoji.Check()),
			SubtaskAdded:       fmt.Sprintf("%s Подзадача добавлена к <b>[%%s]</b>: %%s", emoji.Check()),
			CommentAdded:       fmt.Sprintf("%s Комментарий добавлен к <b>[%%s]</b>.", emoji.Check()),
			StatusChangedAlert: "Статус изменен: %s %s",
		},
		View: View{
			UsageHint:  fmt.Sprintf("%s Использование: /view B0 или /view I1", emoji.Info()),
			NotFound:   fmt.Sprintf("%s Задача <code>%%s</code> не найдена.", emoji.Cross()),
			ListHeader: fmt.Sprintf("%s <b>Список задач (%%d)</b>\nФильтр: <code>%%s</code> | <code>%%s</code>\n\nВыберите задачу для просмотра или редактирования:", emoji.Clipboard()),
			PageFormat: "Стр. %d/%d",
		},
		Task: Task{
			Header:         "<b>[%s]</b> %s <b>%s: %s</b>\n\n",
			StatusLabel:    "<b>Статус:</b> %s %s\n",
			AuthorLabel:    "<b>Автор:</b> @%s\n",
			CreatedLabel:   "<b>Создано:</b> %s UTC\n\n",
			DescLabel:      "<b>Описание:</b>\n<blockquote expandable>%s</blockquote>\n\n",
			SubtasksLabel:  "<b>Подзадачи (%d/%d):</b>\n",
			CommentsLabel:  "<b>Комментарии (%d):</b>\n",
			TypeBug:        "Баг",
			TypeIdea:       "Идея",
			StatusNew:      "Новое",
			StatusProgress: "В работе",
			StatusDone:     "Готово",
			StatusRejected: "Отклонено",
		},
		Buttons: Buttons{
			InProgress:   fmt.Sprintf("%s В работу", emoji.Gear()),
			Done:         fmt.Sprintf("%s Готово", emoji.Check()),
			Rejected:     fmt.Sprintf("%s Отклонить", emoji.Cross()),
			EditTitle:    fmt.Sprintf("%s Заголовок", emoji.Pencil()),
			EditDesc:     fmt.Sprintf("%s Описание", emoji.Memo()),
			AddSubtask:   fmt.Sprintf("%s Саб-таск", emoji.Plus()),
			AddComment:   fmt.Sprintf("%s Комментарий", emoji.Messages()),
			OriginalPost: fmt.Sprintf("%s Исходный пост", emoji.Link()),
			Refresh:      fmt.Sprintf("%s Обновить", emoji.Refresh()),
			PrevPage:     fmt.Sprintf("%s Назад", emoji.ArrowL()),
			NextPage:     fmt.Sprintf("Вперед %s", emoji.ArrowR()),
			OpenDM:       fmt.Sprintf("%s Написать боту в ЛС", emoji.Rocket()),
		},
		Filters: Filters{
			AllTypes:    "Все",
			Bugs:        fmt.Sprintf("%s Баги", emoji.Bug()),
			Ideas:       fmt.Sprintf("%s Идеи", emoji.Idea()),
			AllStatuses: "Все статусы",
			New:         fmt.Sprintf("%s Новые", emoji.New()),
			InProgress:  fmt.Sprintf("%s В работе", emoji.Gear()),
			Done:        fmt.Sprintf("%s Готово", emoji.Check()),
		},
		Notifications: Notifications{
			PromptStartDM:      fmt.Sprintf("%s %%s, чтобы бот мог отвечать вам, сначала начните с ним диалог в личных сообщениях.", emoji.Warning()),
			TopicStatusUpdated: fmt.Sprintf("%s <b>Статус задачи [%%s] обновлен!</b>\n\n<b>Новый статус:</b> %%s %%s\n<b>Заголовок:</b> %%s", emoji.Bell()),
			DMStatusUpdated:    fmt.Sprintf("%s <b>Обновление по вашей задаче [%%s]!</b>\n\n<b>Новый статус:</b> %%s %%s\n<b>Заголовок:</b> %%s\n\n<i>(Вы можете отключить уведомления в личку кнопкой ниже)</i>", emoji.Bell()),
		},
	}
}
