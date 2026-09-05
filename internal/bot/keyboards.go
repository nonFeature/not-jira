package bot

import (
	"fmt"

	"not-jira/internal/emoji"
	"not-jira/internal/locales"
	"not-jira/internal/models"

	"github.com/mymmrac/telego"
)

func BuildTaskInlineKeyboard(task *models.Task, userID int64, isAdmin bool, isDev bool, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton
	canManage := isAdmin || (task.AssigneeID != 0 && task.AssigneeID == userID)
	isDevOrAdmin := isAdmin || isDev

	// Subtasks checkboxes row
	for _, sub := range task.Subtasks {
		emojiID := emoji.ID_SUB_EMPTY
		fallback := "⬜️"
		if sub.IsDone {
			emojiID = emoji.ID_SUB_DONE
			fallback = "☑️"
		}
		titleRunes := []rune(sub.Title)
		if len(titleRunes) > 35 {
			sub.Title = string(titleRunes[:32]) + "..."
		}
		callback := "noop"
		if canManage {
			callback = fmt.Sprintf("toggle_sub:%d:%s", sub.ID, task.ID)
		}
		btn := emoji.MakeInlineButton(sub.Title, callback, "", emojiID, fallback, "")
		rows = append(rows, []telego.InlineKeyboardButton{btn})
	}

	// Status action row (GitHub style)
	if task.IsClosed() {
		if canManage {
			var closedRow []telego.InlineKeyboardButton
			closedRow = append(closedRow, emoji.MakeInlineButton(
				l.Buttons.Reopen,
				fmt.Sprintf("reopen:%s", task.ID),
				"",
				emoji.ID_REFRESH,
				"🔄",
				"primary",
			))
			if isAdmin && !task.IsArchived {
				closedRow = append(closedRow, emoji.MakeInlineButton(
					l.Buttons.Archive,
					fmt.Sprintf("archive:%s", task.ID),
					"",
					emoji.ID_BOX,
					"📦",
					"",
				))
			}
			rows = append(rows, closedRow)
		}
	} else {
		// Task is OPEN
		if canManage {
			var openStatusRow []telego.InlineKeyboardButton
			if task.Status == models.StatusNew {
				openStatusRow = append(openStatusRow, emoji.MakeInlineButton(
					l.Buttons.InProgress,
					fmt.Sprintf("set_status:%s:IN_PROGRESS", task.ID),
					"",
					emoji.ID_GEAR,
					"⚙️",
					"primary",
				))
			}
			openStatusRow = append(openStatusRow, emoji.MakeInlineButton(
				l.Buttons.Done,
				fmt.Sprintf("set_status:%s:DONE", task.ID),
				"",
				emoji.ID_CHECK,
				"✅",
				"success",
			))
			openStatusRow = append(openStatusRow, emoji.MakeInlineButton(
				l.Buttons.Rejected,
				fmt.Sprintf("set_status:%s:REJECTED", task.ID),
				"",
				emoji.ID_CROSS,
				"❌",
				"danger",
			))
			rows = append(rows, openStatusRow)
		}
	}

	// Assignee Row
	if task.IsOpen() {
		var assignRow []telego.InlineKeyboardButton
		if task.AssigneeID == 0 {
			if isDevOrAdmin {
				assignRow = append(assignRow, emoji.MakeInlineButton(
					l.Buttons.Claim,
					fmt.Sprintf("claim:%s", task.ID),
					"",
					emoji.ID_USER,
					"👤",
					"primary",
				))
			}
			if isAdmin {
				assignRow = append(assignRow, emoji.MakeInlineButton(
					l.Buttons.Transfer,
					fmt.Sprintf("transfer:%s", task.ID),
					"",
					emoji.ID_USER,
					"👤",
					"",
				))
			}
		} else if canManage {
			assignRow = append(assignRow, emoji.MakeInlineButton(
				l.Buttons.Transfer,
				fmt.Sprintf("transfer:%s", task.ID),
				"",
				emoji.ID_USER,
				"👤",
				"",
			))
		}
		if len(assignRow) > 0 {
			rows = append(rows, assignRow)
		}
	}

	// Edit actions
	if canManage {
		if isAdmin {
			rows = append(rows, []telego.InlineKeyboardButton{
				emoji.MakeInlineButton(l.Buttons.EditTitle, fmt.Sprintf("edit_title:%s", task.ID), "", emoji.ID_PENCIL, "✏️", ""),
				emoji.MakeInlineButton(l.Buttons.EditDesc, fmt.Sprintf("edit_desc:%s", task.ID), "", emoji.ID_MEMO, "📝", ""),
			})
		}
		rows = append(rows, []telego.InlineKeyboardButton{
			emoji.MakeInlineButton(l.Buttons.AddSubtask, fmt.Sprintf("add_sub:%s", task.ID), "", emoji.ID_PLUS, "➕", ""),
			emoji.MakeInlineButton(l.Buttons.AddComment, fmt.Sprintf("add_comm:%s", task.ID), "", emoji.ID_MESSAGES, "💬", ""),
		})
	}

	// Bottom row
	var bottomRow []telego.InlineKeyboardButton
	if task.MessageLink != "" {
		bottomRow = append(bottomRow, emoji.MakeInlineButton(
			l.Buttons.OriginalPost,
			"",
			task.MessageLink,
			emoji.ID_LINK,
			"🔗",
			"",
		))
	}
	bottomRow = append(bottomRow, emoji.MakeInlineButton(
		l.Buttons.Refresh,
		fmt.Sprintf("view:%s", task.ID),
		"",
		emoji.ID_REFRESH,
		"🔄",
		"",
	))
	rows = append(rows, bottomRow)

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildTransferInviteKeyboard(taskID string, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				emoji.MakeInlineButton(l.Buttons.AcceptTask, fmt.Sprintf("accept_assign:%s", taskID), "", emoji.ID_CHECK, "✅", "success"),
				emoji.MakeInlineButton(l.Buttons.RejectTask, fmt.Sprintf("reject_assign:%s", taskID), "", emoji.ID_CROSS, "❌", "danger"),
			},
		},
	})
}

func BuildListKeyboard(tasks []models.Task, currentType, currentStatus string, page, totalPages int, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Type filters: ALL, BUG, IDEA
	typeRow := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(filterLabel(l.Filters.AllTypes, currentType == "ALL"), fmt.Sprintf("list:ALL:%s:0", currentStatus), "", "", "", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.Bugs, currentType == "BUG"), fmt.Sprintf("list:BUG:%s:0", currentStatus), "", emoji.ID_BUG, "🪲", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.Ideas, currentType == "IDEA"), fmt.Sprintf("list:IDEA:%s:0", currentStatus), "", emoji.ID_IDEA, "💡", ""),
	}
	rows = append(rows, typeRow)

	// Status filters: ALL, NEW, IN_PROGRESS, DONE
	statusRow := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(filterLabel(l.Filters.AllStatuses, currentStatus == "ALL"), fmt.Sprintf("list:%s:ALL:0", currentType), "", "", "", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.New, currentStatus == "NEW"), fmt.Sprintf("list:%s:NEW:0", currentType), "", emoji.ID_NEW, "🆕", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.InProgress, currentStatus == "IN_PROGRESS"), fmt.Sprintf("list:%s:IN_PROGRESS:0", currentType), "", emoji.ID_GEAR, "⚙️", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.Done, currentStatus == "DONE"), fmt.Sprintf("list:%s:DONE:0", currentType), "", emoji.ID_CHECK, "✅", ""),
	}
	rows = append(rows, statusRow)

	// Task buttons for the current page
	for _, t := range tasks {
		cleanTitle := emoji.StripCustomEmojis(t.Title)
		titleRunes := []rune(cleanTitle)
		if len(titleRunes) > 32 {
			cleanTitle = string(titleRunes[:29]) + "..."
		}
		btnText := fmt.Sprintf("[%s] %s %s", t.ID, TaskStatusUnicode(t.Status), cleanTitle)
		btn := telego.InlineKeyboardButton{
			Text:         btnText,
			CallbackData: fmt.Sprintf("view:%s", t.ID),
		}
		rows = append(rows, []telego.InlineKeyboardButton{btn})
	}

	// Pagination row
	var navRow []telego.InlineKeyboardButton
	if page > 0 {
		navRow = append(navRow, emoji.MakeInlineButton(
			l.Buttons.PrevPage,
			fmt.Sprintf("list:%s:%s:%d", currentType, currentStatus, page-1),
			"",
			emoji.ID_ARROW_L,
			"⬅️",
			"",
		))
	}

	pageLabel := fmt.Sprintf(l.View.PageFormat, page+1, max(1, totalPages))
	navRow = append(navRow, telego.InlineKeyboardButton{
		Text:         pageLabel,
		CallbackData: "noop",
	})

	if page+1 < totalPages {
		navRow = append(navRow, emoji.MakeInlineButton(
			l.Buttons.NextPage,
			fmt.Sprintf("list:%s:%s:%d", currentType, currentStatus, page+1),
			"",
			emoji.ID_ARROW_R,
			"➡️",
			"",
		))
	}
	rows = append(rows, navRow)

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildSettingsKeyboard(notifyDM bool, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var btn telego.InlineKeyboardButton
	if notifyDM {
		btn = emoji.MakeInlineButton(l.Settings.BtnDisable, "settings:disable", "", emoji.ID_BELL_OFF, "🔕", "danger")
	} else {
		btn = emoji.MakeInlineButton(l.Settings.BtnEnable, "settings:enable", "", emoji.ID_BELL, "🔔", "success")
	}
	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{{btn}},
	})
}

func sanitizeKeyboard(kb *telego.InlineKeyboardMarkup) *telego.InlineKeyboardMarkup {
	if kb == nil {
		return nil
	}
	for i := range kb.InlineKeyboard {
		for j := range kb.InlineKeyboard[i] {
			kb.InlineKeyboard[i][j].Text = emoji.StripCustomEmojis(kb.InlineKeyboard[i][j].Text)
		}
	}
	return kb
}

func filterLabel(name string, active bool) string {
	cleanName := emoji.StripCustomEmojis(name)
	if active {
		return "• " + cleanName + " •"
	}
	return cleanName
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
