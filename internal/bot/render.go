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
		return emoji.Question()
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
	desc := task.Description
	descRunes := []rune(desc)
	if len(descRunes) > 1500 {
		desc = string(descRunes[:1497]) + "..."
	}
	sb.WriteString(fmt.Sprintf(l.Task.DescLabel, html.EscapeString(desc)))

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
			cText := c.Text
			cRunes := []rune(cText)
			if len(cRunes) > 120 {
				cText = string(cRunes[:117]) + "..."
			}
			sb.WriteString(fmt.Sprintf("%s <i>%s</i>: %s\n", commentIcon, html.EscapeString(c.AuthorName), html.EscapeString(cText)))
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

func formatTimeAgo(t time.Time, l *locales.Bundle) string {
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
		return fmt.Sprintf(l.Task.MinutesAgo, mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours < 1 {
			hours = 1
		}
		return fmt.Sprintf(l.Task.HoursAgo, hours)
	}
	days := int(diff.Hours() / 24)
	if days < 1 {
		days = 1
	}
	return fmt.Sprintf(l.Task.DaysAgo, days)
}

func formatRelativeTime(t time.Time, l *locales.Bundle) string {
	return fmt.Sprintf(l.Task.UpdatedLabel, formatTimeAgo(t, l))
}

func RenderTaskHistory(taskID string, entries []models.HistoryEntry, l *locales.Bundle) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(l.History.Title, html.EscapeString(taskID)))
	sb.WriteString("\n\n")

	if len(entries) == 0 {
		sb.WriteString(l.History.Empty)
		return sb.String()
	}

	for _, e := range entries {
		sb.WriteString(renderHistoryEntry(e, l))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(l.History.MetaLine, html.EscapeString(historyAuthor(e.AuthorName)), formatTimeAgo(e.CreatedAt, l)))
		sb.WriteString("\n\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

func renderHistoryEntry(e models.HistoryEntry, l *locales.Bundle) string {
	switch e.Action {
	case models.HistoryActionCreated:
		return fmt.Sprintf(l.History.EventLine, l.History.Created)
	case models.HistoryActionArchive:
		return fmt.Sprintf(l.History.EventLine, l.History.Archived)
	case models.HistoryActionReopen:
		return fmt.Sprintf(l.History.EventLine, l.History.Reopened)
	case models.HistoryActionStatus:
		return fmt.Sprintf(l.History.EntryLine, l.History.FieldStatus,
			TaskStatusName(models.TaskStatus(e.OldValue), l), TaskStatusName(models.TaskStatus(e.NewValue), l))
	case models.HistoryActionPriority:
		return fmt.Sprintf(l.History.EntryLine, l.History.FieldPriority,
			TaskPriorityName(models.TaskPriority(e.OldValue), l), TaskPriorityName(models.TaskPriority(e.NewValue), l))
	case models.HistoryActionAssignee:
		return fmt.Sprintf(l.History.EntryLine, l.History.FieldAssignee,
			historyValue(e.OldValue, l.History.Unassigned), historyValue(e.NewValue, l.History.Unassigned))
	case models.HistoryActionTitle:
		return fmt.Sprintf(l.History.EntryLine, l.History.FieldTitle,
			historyValue(truncateText(e.OldValue, 80), l.History.None), historyValue(truncateText(e.NewValue, 80), l.History.None))
	case models.HistoryActionDesc:
		return fmt.Sprintf(l.History.EntryLine, l.History.FieldDesc,
			historyValue(truncateText(e.OldValue, 80), l.History.None), historyValue(truncateText(e.NewValue, 80), l.History.None))
	case models.HistoryActionLabels:
		return fmt.Sprintf(l.History.EntryLine, l.History.FieldLabels,
			historyValue(e.OldValue, l.History.None), historyValue(e.NewValue, l.History.None))
	default:
		return fmt.Sprintf(l.History.EventLine, html.EscapeString(e.Action))
	}
}

func historyValue(raw, emptyLabel string) string {
	if strings.TrimSpace(raw) == "" {
		return "<i>" + html.EscapeString(emptyLabel) + "</i>"
	}
	return html.EscapeString(raw)
}

func historyAuthor(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "—"
	}
	return "@" + strings.TrimPrefix(name, "@")
}

func truncateText(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "..."
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
