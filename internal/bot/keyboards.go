package bot

import (
	"fmt"
	"strings"

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
			if task.AssigneeID == userID {
				assignRow = append(assignRow, emoji.MakeInlineButton(
					l.Buttons.Unclaim,
					fmt.Sprintf("unclaim:%s", task.ID),
					"",
					emoji.ID_USER,
					"👤",
					"danger",
				))
			}
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
			emoji.MakeInlineButton(l.Buttons.Priority, fmt.Sprintf("pick_priority:%s", task.ID), "", emoji.ID_PRIORITY, "⚡️", ""),
			emoji.MakeInlineButton(l.Buttons.Labels, fmt.Sprintf("edit_labels:%s", task.ID), "", emoji.ID_TAG, "🏷", ""),
		})
		rows = append(rows, []telego.InlineKeyboardButton{
			emoji.MakeInlineButton(l.Buttons.AddSubtask, fmt.Sprintf("manage_sub:%s", task.ID), "", emoji.ID_PLUS, "➕", ""),
			emoji.MakeInlineButton(l.Buttons.AddComment, fmt.Sprintf("manage_comm:%s", task.ID), "", emoji.ID_MESSAGES, "💬", ""),
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

func BuildSubtasksManageKeyboard(task *models.Task, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	for i, sub := range task.Subtasks {
		cleanTitle := emoji.StripCustomEmojis(sub.Title)
		titleRunes := []rune(cleanTitle)
		if len(titleRunes) > 40 {
			cleanTitle = string(titleRunes[:37]) + "..."
		}
		emojiID := emoji.ID_SUB_EMPTY
		fallback := "⬜️"
		if sub.IsDone {
			emojiID = emoji.ID_SUB_DONE
			fallback = "☑️"
		}
		label := fmt.Sprintf("%d. %s", i+1, cleanTitle)
		btn := emoji.MakeInlineButton(
			label,
			fmt.Sprintf("sub_item:%d:%s", sub.ID, task.ID),
			"",
			emojiID,
			fallback,
			"",
		)
		rows = append(rows, []telego.InlineKeyboardButton{btn})
	}

	// Action row: [➕ Добавить] [🗑 Очистить все]
	var actionRow []telego.InlineKeyboardButton
	actionRow = append(actionRow, emoji.MakeInlineButton(
		l.Buttons.Add,
		fmt.Sprintf("add_sub_prompt:%s", task.ID),
		"",
		emoji.ID_PLUS,
		"➕",
		"primary",
	))
	if len(task.Subtasks) > 0 {
		actionRow = append(actionRow, emoji.MakeInlineButton(
			l.Buttons.ClearAll,
			fmt.Sprintf("clear_subs:%s", task.ID),
			"",
			emoji.ID_TRASH,
			"🗑",
			"danger",
		))
	}
	rows = append(rows, actionRow)

	// Back row: [← Назад]
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(l.Buttons.Back, fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", ""),
	})

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildSubtaskItemKeyboard(taskID string, sub *models.Subtask, itemNum int, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	cleanTitle := emoji.StripCustomEmojis(sub.Title)
	titleRunes := []rune(cleanTitle)
	if len(titleRunes) > 35 {
		cleanTitle = string(titleRunes[:32]) + "..."
	}
	emojiID := emoji.ID_SUB_EMPTY
	fallback := "⬜️"
	if sub.IsDone {
		emojiID = emoji.ID_SUB_DONE
		fallback = "☑️"
	}
	label := fmt.Sprintf("%d. %s", itemNum, cleanTitle)
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(
			label,
			"noop",
			"",
			emojiID,
			fallback,
			"",
		),
	})

	// Action row: [✏️ Изменить] [🗑 Удалить]
	rowAction := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(
			l.Buttons.EditAction,
			fmt.Sprintf("edit_sub:%d:%s", sub.ID, taskID),
			"",
			emoji.ID_PENCIL,
			"✏️",
			"primary",
		),
		emoji.MakeInlineButton(
			l.Buttons.DeleteAction,
			fmt.Sprintf("del_sub:%d:%s", sub.ID, taskID),
			"",
			emoji.ID_TRASH,
			"🗑",
			"danger",
		),
	}
	rows = append(rows, rowAction)

	// Back row: [← К подзадачам]
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(
			l.Buttons.BackToSubtasks,
			fmt.Sprintf("manage_sub:%s", taskID),
			"",
			emoji.ID_ARROW_L,
			"⬅️",
			"",
		),
	})

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildCommentsManageKeyboard(task *models.Task, userID int64, isAdmin bool, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	for i, comm := range task.Comments {
		cleanText := emoji.StripCustomEmojis(comm.Text)
		cleanText = strings.ReplaceAll(cleanText, "\n", " ")
		textRunes := []rune(cleanText)
		if len(textRunes) > 35 {
			cleanText = string(textRunes[:32]) + "..."
		}
		label := fmt.Sprintf("%d. @%s: %s", i+1, comm.AuthorName, cleanText)
		btn := emoji.MakeInlineButton(
			label,
			fmt.Sprintf("comm_item:%d:%s", comm.ID, task.ID),
			"",
			emoji.ID_MESSAGES,
			"💬",
			"",
		)
		rows = append(rows, []telego.InlineKeyboardButton{btn})
	}

	// Action row: [💬 Добавить] [🗑 Очистить все] (if admin)
	var actionRow []telego.InlineKeyboardButton
	actionRow = append(actionRow, emoji.MakeInlineButton(
		l.Buttons.Add,
		fmt.Sprintf("add_comm_prompt:%s", task.ID),
		"",
		emoji.ID_MESSAGES,
		"💬",
		"primary",
	))
	if isAdmin && len(task.Comments) > 0 {
		actionRow = append(actionRow, emoji.MakeInlineButton(
			l.Buttons.ClearAll,
			fmt.Sprintf("clear_comms:%s", task.ID),
			"",
			emoji.ID_TRASH,
			"🗑",
			"danger",
		))
	}
	rows = append(rows, actionRow)

	// Back row: [← Назад]
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(l.Buttons.Back, fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", ""),
	})

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildCommentItemKeyboard(taskID string, comm *models.Comment, itemNum int, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	cleanText := emoji.StripCustomEmojis(comm.Text)
	cleanText = strings.ReplaceAll(cleanText, "\n", " ")
	textRunes := []rune(cleanText)
	if len(textRunes) > 30 {
		cleanText = string(textRunes[:27]) + "..."
	}

	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(
			fmt.Sprintf("%d. @%s: %s", itemNum, comm.AuthorName, cleanText),
			"noop",
			"",
			emoji.ID_MESSAGES,
			"💬",
			"",
		),
	})

	// Action row: [✏️ Изменить] [🗑 Удалить]
	rowAction := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(
			l.Buttons.EditAction,
			fmt.Sprintf("edit_comm:%d:%s", comm.ID, taskID),
			"",
			emoji.ID_PENCIL,
			"✏️",
			"primary",
		),
		emoji.MakeInlineButton(
			l.Buttons.DeleteAction,
			fmt.Sprintf("del_comm:%d:%s", comm.ID, taskID),
			"",
			emoji.ID_TRASH,
			"🗑",
			"danger",
		),
	}
	rows = append(rows, rowAction)

	// Back row: [← К комментариям]
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(
			l.Buttons.BackToComments,
			fmt.Sprintf("manage_comm:%s", taskID),
			"",
			emoji.ID_ARROW_L,
			"⬅️",
			"",
		),
	})

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

func BuildListKeyboard(tasks []models.Task, currentType, currentStatus, currentTag string, page, totalPages int, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	if currentTag == "" {
		currentTag = "ALL"
	}

	// Type filters: ALL, BUG, IDEA
	typeRow := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(filterLabel(l.Filters.AllTypes, currentType == "ALL"), fmt.Sprintf("list:ALL:%s:%s:0", currentStatus, currentTag), "", "", "", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.Bugs, currentType == "BUG"), fmt.Sprintf("list:BUG:%s:%s:0", currentStatus, currentTag), "", emoji.ID_BUG, "🪲", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.Ideas, currentType == "IDEA"), fmt.Sprintf("list:IDEA:%s:%s:0", currentStatus, currentTag), "", emoji.ID_IDEA, "💡", ""),
	}
	rows = append(rows, typeRow)

	// Status filters: ALL, NEW, IN_PROGRESS, DONE
	statusRow := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(filterLabel(l.Filters.AllStatuses, currentStatus == "ALL"), fmt.Sprintf("list:%s:ALL:%s:0", currentType, currentTag), "", "", "", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.New, currentStatus == "NEW"), fmt.Sprintf("list:%s:NEW:%s:0", currentType, currentTag), "", emoji.ID_NEW, "🆕", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.InProgress, currentStatus == "IN_PROGRESS"), fmt.Sprintf("list:%s:IN_PROGRESS:%s:0", currentType, currentTag), "", emoji.ID_GEAR, "⚙️", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.Done, currentStatus == "DONE"), fmt.Sprintf("list:%s:DONE:%s:0", currentType, currentTag), "", emoji.ID_CHECK, "✅", ""),
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
			fmt.Sprintf("list:%s:%s:%s:%d", currentType, currentStatus, currentTag, page-1),
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
			fmt.Sprintf("list:%s:%s:%s:%d", currentType, currentStatus, currentTag, page+1),
			"",
			emoji.ID_ARROW_R,
			"➡️",
			"",
		))
	}
	rows = append(rows, navRow)

	// Bottom row with Tag filter and My Tasks
	tagBtnText := l.Filters.AllTags
	if currentTag != "" && currentTag != "ALL" {
		tagBtnText = "#" + strings.TrimPrefix(currentTag, "#")
	}

	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(tagBtnText, fmt.Sprintf("list_tags:%s:%s:%s:%d", currentType, currentStatus, currentTag, page), "", emoji.ID_TAG, "🏷", ""),
		emoji.MakeInlineButton(l.Buttons.MyTasks, "my:assigned:0", "", emoji.ID_USER, "👤", ""),
	})

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildListTagsKeyboard(currentType, currentStatus, currentTag string, page int, availableTags []string, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// All tags option
	allLabel := l.Filters.AllTags
	if currentTag == "" || currentTag == "ALL" {
		allLabel = "• " + allLabel + " •"
	}
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(allLabel, fmt.Sprintf("list:%s:%s:ALL:0", currentType, currentStatus), "", emoji.ID_TAG, "🏷", ""),
	})

	// 2 tags per row
	var tagRow []telego.InlineKeyboardButton
	for _, tag := range availableTags {
		cleanTag := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if cleanTag == "" {
			continue
		}
		label := "#" + cleanTag
		if strings.EqualFold(cleanTag, currentTag) {
			label = "• " + label + " •"
		}
		tagRow = append(tagRow, emoji.MakeInlineButton(
			label,
			fmt.Sprintf("list:%s:%s:%s:0", currentType, currentStatus, cleanTag),
			"",
			emoji.ID_TAG,
			"🏷",
			"",
		))
		if len(tagRow) == 2 {
			rows = append(rows, tagRow)
			tagRow = nil
		}
	}
	if len(tagRow) > 0 {
		rows = append(rows, tagRow)
	}

	// Back button
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(l.Buttons.Back, fmt.Sprintf("list:%s:%s:%s:%d", currentType, currentStatus, currentTag, page), "", emoji.ID_ARROW_L, "⬅️", ""),
	})

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildPriorityPickerKeyboard(taskID string, current models.TaskPriority, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	mark := func(p models.TaskPriority, label string) string {
		if p == current {
			return "• " + label + " •"
		}
		return label
	}

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				emoji.MakeInlineButton(mark(models.PriorityP0, l.Task.P0), fmt.Sprintf("set_priority:%s:P0", taskID), "", emoji.ID_P0, "🔴", "danger"),
				emoji.MakeInlineButton(mark(models.PriorityP1, l.Task.P1), fmt.Sprintf("set_priority:%s:P1", taskID), "", emoji.ID_P1, "🟡", ""),
			},
			{
				emoji.MakeInlineButton(mark(models.PriorityP2, l.Task.P2), fmt.Sprintf("set_priority:%s:P2", taskID), "", emoji.ID_P2, "🔵", "primary"),
				emoji.MakeInlineButton(mark(models.PriorityP3, l.Task.P3), fmt.Sprintf("set_priority:%s:P3", taskID), "", emoji.ID_P3, "⚪️", ""),
			},
			{
				emoji.MakeInlineButton(l.Buttons.Back, fmt.Sprintf("view:%s", taskID), "", emoji.ID_ARROW_L, "⬅️", ""),
			},
		},
	})
}

func BuildTaskLabelsKeyboard(task *models.Task, availableTags []string, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	var tagRow []telego.InlineKeyboardButton
	for _, rawTag := range availableTags {
		cleanTag := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(rawTag, "#")))
		if cleanTag == "" {
			continue
		}
		btnLabel := "#" + cleanTag
		has := task.HasLabel(cleanTag)
		if has {
			btnLabel = "✅ #" + cleanTag
		}
		btnType := ""
		if has {
			btnType = "primary"
		}
		tagRow = append(tagRow, emoji.MakeInlineButton(
			btnLabel,
			fmt.Sprintf("toggle_label:%s:%s", task.ID, cleanTag),
			"",
			emoji.ID_TAG,
			"🏷",
			btnType,
		))
		if len(tagRow) == 2 {
			rows = append(rows, tagRow)
			tagRow = nil
		}
	}
	if len(tagRow) > 0 {
		rows = append(rows, tagRow)
	}

	// Action row: [✏️ Ввести вручную] [🗑 Очистить]
	actionRow := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(l.Buttons.EditLabelsManual, fmt.Sprintf("edit_labels_manual:%s", task.ID), "", emoji.ID_PENCIL, "✏️", ""),
		emoji.MakeInlineButton(l.Buttons.ClearLabels, fmt.Sprintf("clear_labels:%s", task.ID), "", emoji.ID_TRASH, "🗑", "danger"),
	}
	rows = append(rows, actionRow)

	// Back button
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(l.Buttons.Back, fmt.Sprintf("view:%s", task.ID), "", emoji.ID_ARROW_L, "⬅️", ""),
	})

	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func BuildCancelKeyboard(taskID string, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	callbackData := "cancel_fsm"
	if taskID != "" {
		callbackData = "cancel_fsm:" + taskID
	}
	return sanitizeKeyboard(&telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				emoji.MakeInlineButton(l.Buttons.Cancel, callbackData, "", emoji.ID_CROSS, "❌", "danger"),
			},
		},
	})
}

func BuildMyTasksKeyboard(tasks []models.Task, currentTab string, page, totalPages int, l *locales.Bundle) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Tab selection row: [ На мне ] | [ Созданные мной ]
	tabRow := []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(filterLabel(l.Filters.MyAssigned, currentTab == "assigned"), "my:assigned:0", "", emoji.ID_USER, "👤", ""),
		emoji.MakeInlineButton(filterLabel(l.Filters.MyCreated, currentTab == "created"), "my:created:0", "", emoji.ID_WRITE, "✍️", ""),
	}
	rows = append(rows, tabRow)

	// Task buttons for current page
	for _, t := range tasks {
		cleanTitle := emoji.StripCustomEmojis(t.Title)
		titleRunes := []rune(cleanTitle)
		if len(titleRunes) > 30 {
			cleanTitle = string(titleRunes[:27]) + "..."
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
			fmt.Sprintf("my:%s:%d", currentTab, page-1),
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
			fmt.Sprintf("my:%s:%d", currentTab, page+1),
			"",
			emoji.ID_ARROW_R,
			"➡️",
			"",
		))
	}
	rows = append(rows, navRow)

	// Return to full list button
	rows = append(rows, []telego.InlineKeyboardButton{
		emoji.MakeInlineButton(l.Filters.AllStatuses, "list:ALL:ALL:ALL:0", "", emoji.ID_CLIPBOARD, "📋", ""),
	})

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
