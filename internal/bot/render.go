package bot

import (
	"fmt"
	"html"
	"strings"

	"not-jira/internal/emoji"
	"not-jira/internal/locales"
	"not-jira/internal/models"
)

func TaskTypeEmoji(t models.TaskType) string {
	if t == models.TaskTypeIdea {
		return emoji.Idea()
	}
	return emoji.Bug()
}

func TaskTypeName(t models.TaskType, l *locales.Bundle) string {
	if t == models.TaskTypeIdea {
		return l.Task.TypeIdea
	}
	return l.Task.TypeBug
}

func TaskStatusEmoji(s models.TaskStatus) string {
	switch s {
	case models.StatusNew:
		return emoji.New()
	case models.StatusInProgress:
		return emoji.Gear()
	case models.StatusDone:
		return emoji.Check()
	case models.StatusRejected:
		return emoji.Cross()
	default:
		return "❓"
	}
}

func TaskStatusName(s models.TaskStatus, l *locales.Bundle) string {
	switch s {
	case models.StatusNew:
		return l.Task.StatusNew
	case models.StatusInProgress:
		return l.Task.StatusProgress
	case models.StatusDone:
		return l.Task.StatusDone
	case models.StatusRejected:
		return l.Task.StatusRejected
	default:
		return string(s)
	}
}

func RenderTaskCard(task *models.Task, l *locales.Bundle) string {
	var sb strings.Builder

	// Header: [B0] 🐛 Bug/Баг: Title
	sb.WriteString(fmt.Sprintf(l.Task.Header,
		task.ID, TaskTypeEmoji(task.Type), TaskTypeName(task.Type, l), html.EscapeString(task.Title)))

	// Status & Metadata
	sb.WriteString(fmt.Sprintf(l.Task.StatusLabel, TaskStatusEmoji(task.Status), TaskStatusName(task.Status, l)))
	if task.AuthorUsername != "" {
		sb.WriteString(fmt.Sprintf(l.Task.AuthorLabel, html.EscapeString(task.AuthorUsername)))
	}
	sb.WriteString(fmt.Sprintf(l.Task.CreatedLabel, task.CreatedAt.Format("02.01.2006 15:04")))

	// Description inside an expandable blockquote (Bot API 7.x)
	sb.WriteString(fmt.Sprintf(l.Task.DescLabel, html.EscapeString(task.Description)))

	// Subtasks (if any)
	if len(task.Subtasks) > 0 {
		doneCount := 0
		for _, s := range task.Subtasks {
			if s.IsDone {
				doneCount++
			}
		}
		sb.WriteString(fmt.Sprintf(l.Task.SubtasksLabel, doneCount, len(task.Subtasks)))
		for _, s := range task.Subtasks {
			check := emoji.SubEmpty()
			if s.IsDone {
				check = emoji.SubDone()
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", check, html.EscapeString(s.Title)))
		}
		sb.WriteString("\n")
	}

	// Comments (if any)
	if len(task.Comments) > 0 {
		sb.WriteString(fmt.Sprintf(l.Task.CommentsLabel, len(task.Comments)))
		// Show up to 5 last comments
		start := 0
		if len(task.Comments) > 5 {
			start = len(task.Comments) - 5
		}
		commentIcon := emoji.Messages()
		for _, c := range task.Comments[start:] {
			sb.WriteString(fmt.Sprintf("%s <i>%s</i>: %s\n", commentIcon, html.EscapeString(c.AuthorName), html.EscapeString(c.Text)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func RenderTaskListHeader(totalCount int, filterType, filterStatus string, l *locales.Bundle) string {
	typeDesc := l.Filters.AllTypes
	if filterType == "BUG" {
		typeDesc = l.Filters.Bugs
	} else if filterType == "IDEA" {
		typeDesc = l.Filters.Ideas
	}

	statusDesc := l.Filters.AllStatuses
	if filterStatus != "ALL" {
		s := models.TaskStatus(filterStatus)
		statusDesc = TaskStatusEmoji(s) + " " + TaskStatusName(s, l)
	}

	return fmt.Sprintf(l.View.ListHeader, totalCount, typeDesc, statusDesc)
}
