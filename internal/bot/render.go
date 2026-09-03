package bot

import (
	"fmt"
	"html"
	"strings"

	"not-jira/internal/models"
)

func RenderTaskCard(task *models.Task) string {
	var sb strings.Builder

	// Header: [B0] 🐛 Баг: Заголовок
	sb.WriteString(fmt.Sprintf("<b>[%s]</b> %s <b>%s: %s</b>\n\n",
		task.ID, task.Type.Emoji(), task.Type.Russian(), html.EscapeString(task.Title)))

	// Status & Metadata
	sb.WriteString(fmt.Sprintf("<b>Статус:</b> %s %s\n", task.Status.Emoji(), task.Status.Russian()))
	if task.AuthorUsername != "" {
		sb.WriteString(fmt.Sprintf("<b>Автор:</b> @%s\n", html.EscapeString(task.AuthorUsername)))
	}
	sb.WriteString(fmt.Sprintf("<b>Создано:</b> %s UTC\n\n", task.CreatedAt.Format("02.01.2006 15:04")))

	// Description inside an expandable blockquote (Bot API 7.x)
	sb.WriteString("<b>Описание:</b>\n")
	sb.WriteString(fmt.Sprintf("<blockquote expandable>%s</blockquote>\n\n", html.EscapeString(task.Description)))

	// Subtasks (if any)
	if len(task.Subtasks) > 0 {
		doneCount := 0
		for _, s := range task.Subtasks {
			if s.IsDone {
				doneCount++
			}
		}
		sb.WriteString(fmt.Sprintf("<b>Подзадачи (%d/%d):</b>\n", doneCount, len(task.Subtasks)))
		for _, s := range task.Subtasks {
			check := "⬜️"
			if s.IsDone {
				check = "☑️"
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", check, html.EscapeString(s.Title)))
		}
		sb.WriteString("\n")
	}

	// Comments (if any)
	if len(task.Comments) > 0 {
		sb.WriteString(fmt.Sprintf("<b>Комментарии (%d):</b>\n", len(task.Comments)))
		// Show up to 5 last comments
		start := 0
		if len(task.Comments) > 5 {
			start = len(task.Comments) - 5
		}
		for _, c := range task.Comments[start:] {
			sb.WriteString(fmt.Sprintf("💬 <i>%s</i>: %s\n", html.EscapeString(c.AuthorName), html.EscapeString(c.Text)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func RenderTaskListHeader(totalCount int, filterType, filterStatus string) string {
	typeDesc := "Все типы"
	if filterType == "BUG" {
		typeDesc = "🐛 Баги"
	} else if filterType == "IDEA" {
		typeDesc = "💡 Идеи"
	}

	statusDesc := "Все статусы"
	if filterStatus != "ALL" {
		statusDesc = models.TaskStatus(filterStatus).Emoji() + " " + models.TaskStatus(filterStatus).Russian()
	}

	return fmt.Sprintf("📋 <b>Список задач (%d)</b>\nФильтр: <code>%s</code> | <code>%s</code>\n\nВыберите задачу для просмотра или редактирования:",
		totalCount, typeDesc, statusDesc)
}
