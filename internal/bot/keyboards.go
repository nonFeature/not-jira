package bot

import (
	"fmt"

	"not-jira/internal/models"

	"github.com/mymmrac/telego"
)

func BuildTaskInlineKeyboard(task *models.Task, isAdmin bool) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Subtasks checkboxes row
	for _, sub := range task.Subtasks {
		icon := "⬜️"
		if sub.IsDone {
			icon = "☑️"
		}
		btnText := fmt.Sprintf("%s %s", icon, sub.Title)
		if len(btnText) > 40 {
			btnText = btnText[:37] + "..."
		}
		rows = append(rows, []telego.InlineKeyboardButton{
			{
				Text:         btnText,
				CallbackData: fmt.Sprintf("toggle_sub:%d:%s", sub.ID, task.ID),
			},
		})
	}

	// Status action row
	var statusRow []telego.InlineKeyboardButton
	if task.Status != models.StatusInProgress {
		statusRow = append(statusRow, telego.InlineKeyboardButton{
			Text:         "⚙️ В работу",
			CallbackData: fmt.Sprintf("set_status:%s:IN_PROGRESS", task.ID),
		})
	}
	if task.Status != models.StatusDone {
		statusRow = append(statusRow, telego.InlineKeyboardButton{
			Text:         "✅ Готово",
			CallbackData: fmt.Sprintf("set_status:%s:DONE", task.ID),
		})
	}
	if task.Status != models.StatusRejected {
		statusRow = append(statusRow, telego.InlineKeyboardButton{
			Text:         "❌ Отклонить",
			CallbackData: fmt.Sprintf("set_status:%s:REJECTED", task.ID),
		})
	}
	if len(statusRow) > 0 {
		rows = append(rows, statusRow)
	}

	// GitHub Issues edit actions (admins only)
	if isAdmin {
		rows = append(rows, []telego.InlineKeyboardButton{
			{Text: "✏️ Заголовок", CallbackData: fmt.Sprintf("edit_title:%s", task.ID)},
			{Text: "📝 Описание", CallbackData: fmt.Sprintf("edit_desc:%s", task.ID)},
		})
		rows = append(rows, []telego.InlineKeyboardButton{
			{Text: "➕ Саб-таск", CallbackData: fmt.Sprintf("add_sub:%s", task.ID)},
			{Text: "💬 Комментарий", CallbackData: fmt.Sprintf("add_comm:%s", task.ID)},
		})
	}

	// Link to original message (if available) & Refresh
	var bottomRow []telego.InlineKeyboardButton
	if task.MessageLink != "" {
		bottomRow = append(bottomRow, telego.InlineKeyboardButton{
			Text: "🔗 Исходный пост",
			URL:  task.MessageLink,
		})
	}
	bottomRow = append(bottomRow, telego.InlineKeyboardButton{
		Text:         "🔄 Обновить",
		CallbackData: fmt.Sprintf("view:%s", task.ID),
	})
	rows = append(rows, bottomRow)

	return &telego.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func BuildListKeyboard(tasks []models.Task, currentType, currentStatus string, page, totalPages int) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Type filters: ALL, BUG, IDEA
	typeRow := []telego.InlineKeyboardButton{
		{
			Text:         filterLabel("Все", currentType == "ALL"),
			CallbackData: fmt.Sprintf("list:ALL:%s:0", currentStatus),
		},
		{
			Text:         filterLabel("🐛 Баги", currentType == "BUG"),
			CallbackData: fmt.Sprintf("list:BUG:%s:0", currentStatus),
		},
		{
			Text:         filterLabel("💡 Идеи", currentType == "IDEA"),
			CallbackData: fmt.Sprintf("list:IDEA:%s:0", currentStatus),
		},
	}
	rows = append(rows, typeRow)

	// Status filters: ALL, NEW, IN_PROGRESS, DONE
	statusRow := []telego.InlineKeyboardButton{
		{
			Text:         filterLabel("Все статусы", currentStatus == "ALL"),
			CallbackData: fmt.Sprintf("list:%s:ALL:0", currentType),
		},
		{
			Text:         filterLabel("🆕 Новые", currentStatus == "NEW"),
			CallbackData: fmt.Sprintf("list:%s:NEW:0", currentType),
		},
		{
			Text:         filterLabel("⚙️ В работе", currentStatus == "IN_PROGRESS"),
			CallbackData: fmt.Sprintf("list:%s:IN_PROGRESS:0", currentType),
		},
		{
			Text:         filterLabel("✅ Готово", currentStatus == "DONE"),
			CallbackData: fmt.Sprintf("list:%s:DONE:0", currentType),
		},
	}
	rows = append(rows, statusRow)

	// Task buttons for the current page
	for _, t := range tasks {
		text := fmt.Sprintf("[%s] %s %s", t.ID, t.Status.Emoji(), t.Title)
		if len(text) > 42 {
			text = text[:39] + "..."
		}
		rows = append(rows, []telego.InlineKeyboardButton{
			{
				Text:         text,
				CallbackData: fmt.Sprintf("view:%s", t.ID),
			},
		})
	}

	// Pagination row
	var navRow []telego.InlineKeyboardButton
	if page > 0 {
		navRow = append(navRow, telego.InlineKeyboardButton{
			Text:         "⬅️ Назад",
			CallbackData: fmt.Sprintf("list:%s:%s:%d", currentType, currentStatus, page-1),
		})
	}

	pageLabel := fmt.Sprintf("Стр. %d/%d", page+1, max(1, totalPages))
	navRow = append(navRow, telego.InlineKeyboardButton{
		Text:         pageLabel,
		CallbackData: "noop",
	})

	if page+1 < totalPages {
		navRow = append(navRow, telego.InlineKeyboardButton{
			Text:         "Вперед ➡️",
			CallbackData: fmt.Sprintf("list:%s:%s:%d", currentType, currentStatus, page+1),
		})
	}
	rows = append(rows, navRow)

	return &telego.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func BuildSettingsKeyboard(notifyDM bool) *telego.InlineKeyboardMarkup {
	text := "🔕 Отключить уведы в ЛС"
	if !notifyDM {
		text = "🔔 Включить уведы в ЛС"
	}
	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				{Text: text, CallbackData: "toggle_notify_dm"},
			},
		},
	}
}

func filterLabel(name string, active bool) string {
	if active {
		return "• " + name + " •"
	}
	return name
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
