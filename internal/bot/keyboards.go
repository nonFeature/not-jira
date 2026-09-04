package bot

import (
	"fmt"

	"not-jira/internal/locales"
	"not-jira/internal/models"

	"github.com/mymmrac/telego"
)

func BuildTaskInlineKeyboard(task *models.Task, isAdmin bool, l *locales.Bundle) *telego.InlineKeyboardMarkup {
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
			Text:         l.Buttons.InProgress,
			CallbackData: fmt.Sprintf("set_status:%s:IN_PROGRESS", task.ID),
		})
	}
	if task.Status != models.StatusDone {
		statusRow = append(statusRow, telego.InlineKeyboardButton{
			Text:         l.Buttons.Done,
			CallbackData: fmt.Sprintf("set_status:%s:DONE", task.ID),
		})
	}
	if task.Status != models.StatusRejected {
		statusRow = append(statusRow, telego.InlineKeyboardButton{
			Text:         l.Buttons.Rejected,
			CallbackData: fmt.Sprintf("set_status:%s:REJECTED", task.ID),
		})
	}
	if len(statusRow) > 0 {
		rows = append(rows, statusRow)
	}

	// GitHub Issues edit actions (admins only)
	if isAdmin {
		rows = append(rows, []telego.InlineKeyboardButton{
			{Text: l.Buttons.EditTitle, CallbackData: fmt.Sprintf("edit_title:%s", task.ID)},
			{Text: l.Buttons.EditDesc, CallbackData: fmt.Sprintf("edit_desc:%s", task.ID)},
		})
		rows = append(rows, []telego.InlineKeyboardButton{
			{Text: l.Buttons.AddSubtask, CallbackData: fmt.Sprintf("add_sub:%s", task.ID)},
			{Text: l.Buttons.AddComment, CallbackData: fmt.Sprintf("add_comm:%s", task.ID)},
		})
	}

	// Link to original message (if available) & Refresh
	var bottomRow []telego.InlineKeyboardButton
	if task.MessageLink != "" {
		bottomRow = append(bottomRow, telego.InlineKeyboardButton{
			Text: l.Buttons.OriginalPost,
			URL:  task.MessageLink,
		})
	}
	bottomRow = append(bottomRow, telego.InlineKeyboardButton{
		Text:         l.Buttons.Refresh,
		CallbackData: fmt.Sprintf("view:%s", task.ID),
	})
	rows = append(rows, bottomRow)

	return &telego.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func BuildListKeyboard(tasks []models.Task, currentType, currentStatus string, page, totalPages int, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Type filters: ALL, BUG, IDEA
	typeRow := []telego.InlineKeyboardButton{
		{
			Text:         filterLabel(l.Filters.AllTypes, currentType == "ALL"),
			CallbackData: fmt.Sprintf("list:ALL:%s:0", currentStatus),
		},
		{
			Text:         filterLabel(l.Filters.Bugs, currentType == "BUG"),
			CallbackData: fmt.Sprintf("list:BUG:%s:0", currentStatus),
		},
		{
			Text:         filterLabel(l.Filters.Ideas, currentType == "IDEA"),
			CallbackData: fmt.Sprintf("list:IDEA:%s:0", currentStatus),
		},
	}
	rows = append(rows, typeRow)

	// Status filters: ALL, NEW, IN_PROGRESS, DONE
	statusRow := []telego.InlineKeyboardButton{
		{
			Text:         filterLabel(l.Filters.AllStatuses, currentStatus == "ALL"),
			CallbackData: fmt.Sprintf("list:%s:ALL:0", currentType),
		},
		{
			Text:         filterLabel(l.Filters.New, currentStatus == "NEW"),
			CallbackData: fmt.Sprintf("list:%s:NEW:0", currentType),
		},
		{
			Text:         filterLabel(l.Filters.InProgress, currentStatus == "IN_PROGRESS"),
			CallbackData: fmt.Sprintf("list:%s:IN_PROGRESS:0", currentType),
		},
		{
			Text:         filterLabel(l.Filters.Done, currentStatus == "DONE"),
			CallbackData: fmt.Sprintf("list:%s:DONE:0", currentType),
		},
	}
	rows = append(rows, statusRow)

	// Task buttons for the current page
	for _, t := range tasks {
		text := fmt.Sprintf("[%s] %s %s", t.ID, TaskStatusEmoji(t.Status), t.Title)
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
			Text:         l.Buttons.PrevPage,
			CallbackData: fmt.Sprintf("list:%s:%s:%d", currentType, currentStatus, page-1),
		})
	}

	pageLabel := fmt.Sprintf(l.View.PageFormat, page+1, max(1, totalPages))
	navRow = append(navRow, telego.InlineKeyboardButton{
		Text:         pageLabel,
		CallbackData: "noop",
	})

	if page+1 < totalPages {
		navRow = append(navRow, telego.InlineKeyboardButton{
			Text:         l.Buttons.NextPage,
			CallbackData: fmt.Sprintf("list:%s:%s:%d", currentType, currentStatus, page+1),
		})
	}
	rows = append(rows, navRow)

	return &telego.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func BuildSettingsKeyboard(notifyDM bool, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	text := l.Settings.BtnDisable
	if !notifyDM {
		text = l.Settings.BtnEnable
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
