package bot

import (
	"fmt"
	"html"
	"strings"
	"time"

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

// TaskStatusUnicode returns pure Unicode emoji for button texts (never uses <tg-emoji>).
func TaskStatusUnicode(s models.TaskStatus) string {
	switch s {
	case models.StatusNew:
		return "🆕"
	case models.StatusInProgress:
		return "⚙️"
	case models.StatusDone:
		return "✅"
	case models.StatusRejected:
		return "❌"
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

func TaskPriorityEmoji(p models.TaskPriority) string {
	switch p {
	case models.PriorityP0:
		return emoji.P0()
	case models.PriorityP1:
		return emoji.P1()
	case models.PriorityP2:
		return emoji.P2()
	case models.PriorityP3:
		return emoji.P3()
	default:
		return emoji.P2()
	}
}

func TaskPriorityName(p models.TaskPriority, l *locales.Bundle) string {
	switch p {
	case models.PriorityP0:
		return l.Task.P0
	case models.PriorityP1:
		return l.Task.P1
	case models.PriorityP2:
		return l.Task.P2
	case models.PriorityP3:
		return l.Task.P3
	default:
		return l.Task.P2
	}
}

func RenderTaskCard(task *models.Task, l *locales.Bundle) string {
	var sb strings.Builder

	// Header: [B0] 🐛 Bug/Баг: Title
	sb.WriteString(fmt.Sprintf(l.Task.Header,
		task.ID, TaskTypeEmoji(task.Type), TaskTypeName(task.Type, l), html.EscapeString(task.Title)))

	// Status & Metadata
	statusText := fmt.Sprintf(l.Task.StatusLabel, TaskStatusEmoji(task.Status), TaskStatusName(task.Status, l))
	if task.IsArchived {
		statusText = strings.TrimRight(statusText, "\n") + l.Task.ArchivedBadge + "\n"
	}
	sb.WriteString(statusText)

	priority := task.Priority
	if priority == "" {
		priority = models.PriorityP2
	}
	sb.WriteString(fmt.Sprintf(l.Task.PriorityLabel, TaskPriorityEmoji(priority), TaskPriorityName(priority, l)))

	if len(task.Labels) > 0 {
		formatted := task.FormattedLabels()
		if formatted != "" {
			sb.WriteString(fmt.Sprintf(l.Task.LabelsLabel, html.EscapeString(formatted)))
		}
	}

	if task.AuthorUsername != "" {
		sb.WriteString(fmt.Sprintf(l.Task.AuthorLabel, html.EscapeString(task.AuthorUsername)))
	}
	if task.AssigneeUsername != "" {
		sb.WriteString(fmt.Sprintf(l.Task.AssigneeLabel, html.EscapeString(task.AssigneeUsername)))
	} else {
		sb.WriteString(l.Task.UnassignedLabel)
	}

	updatedTime := task.UpdatedAt
	if updatedTime.IsZero() {
		updatedTime = task.CreatedAt
	}
	sb.WriteString(fmt.Sprintf(l.Task.CreatedLabel, task.CreatedAt.Format("02.01.2006 15:04"), formatRelativeTime(updatedTime, l)))

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
		percent := (doneCount * 100) / len(task.Subtasks)
		bar := renderProgressBar(doneCount, len(task.Subtasks))
		sb.WriteString(fmt.Sprintf(l.Task.SubtasksLabel, doneCount, len(task.Subtasks), bar, percent))
		for _, s := range task.Subtasks {
			check := emoji.SubEmpty()
			title := html.EscapeString(s.Title)
			if s.IsDone {
				check = emoji.SubDone()
				title = "<s>" + title + "</s>"
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", check, title))
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

func renderProgressBar(done, total int) string {
	if total == 0 {
		return ""
	}
	const barLen = 6
	filled := (done * barLen) / total
	if filled > barLen {
		filled = barLen
	}
	empty := barLen - filled
	return strings.Repeat("▓", filled) + strings.Repeat("░", empty)
}

func formatRelativeTime(t time.Time, l *locales.Bundle) string {
	if t.IsZero() {
		return l.Task.JustNow
	}
	diff := time.Since(t)
	if diff < time.Minute {
		return l.Task.JustNow
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins < 1 {
			mins = 1
		}
		return fmt.Sprintf(l.Task.UpdatedLabel, fmt.Sprintf(l.Task.MinutesAgo, mins))
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours < 1 {
			hours = 1
		}
		return fmt.Sprintf(l.Task.UpdatedLabel, fmt.Sprintf(l.Task.HoursAgo, hours))
	}
	days := int(diff.Hours() / 24)
	if days < 1 {
		days = 1
	}
	return fmt.Sprintf(l.Task.UpdatedLabel, fmt.Sprintf(l.Task.DaysAgo, days))
}

func RenderTaskListHeader(totalCount int, filterType, filterStatus, filterTag string, l *locales.Bundle) string {
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

	tagDesc := l.Filters.AllTags
	if filterTag != "" && filterTag != "ALL" {
		tagDesc = "#" + strings.TrimPrefix(filterTag, "#")
	}

	return fmt.Sprintf(l.View.ListHeader, totalCount, typeDesc, statusDesc, tagDesc)
}
