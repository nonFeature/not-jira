package models

import "time"

type TaskType string

const (
	TaskTypeBug  TaskType = "BUG"
	TaskTypeIdea TaskType = "IDEA"
)

func (t TaskType) Prefix() string {
	if t == TaskTypeBug {
		return "B"
	}
	return "I"
}

func (t TaskType) Emoji() string {
	if t == TaskTypeBug {
		return "🐛"
	}
	return "💡"
}

func (t TaskType) Russian() string {
	if t == TaskTypeBug {
		return "Баг"
	}
	return "Идея"
}

type TaskStatus string

const (
	StatusNew        TaskStatus = "NEW"
	StatusInProgress TaskStatus = "IN_PROGRESS"
	StatusDone       TaskStatus = "DONE"
	StatusRejected   TaskStatus = "REJECTED"
)

func (s TaskStatus) Emoji() string {
	switch s {
	case StatusNew:
		return "🆕"
	case StatusInProgress:
		return "⚙️"
	case StatusDone:
		return "✅"
	case StatusRejected:
		return "❌"
	default:
		return "📋"
	}
}

func (s TaskStatus) Russian() string {
	switch s {
	case StatusNew:
		return "Новое"
	case StatusInProgress:
		return "В работе"
	case StatusDone:
		return "Готово"
	case StatusRejected:
		return "Отклонено"
	default:
		return string(s)
	}
}

type Task struct {
	ID               string     `json:"id"` // e.g. "B0", "I3"
	Num              int        `json:"num"`
	Type             TaskType   `json:"type"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Status           TaskStatus `json:"status"`
	ChatID           int64      `json:"chat_id"`
	TopicID          int64      `json:"topic_id"`
	MessageID        int64      `json:"message_id"`
	MessageLink      string     `json:"message_link"`
	AuthorID         int64      `json:"author_id"`
	AuthorUsername   string     `json:"author_username"`
	AssigneeID       int64      `json:"assignee_id,omitempty"`
	AssigneeUsername string     `json:"assignee_username,omitempty"`
	IsArchived       bool       `json:"is_archived"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	Subtasks []Subtask `json:"subtasks,omitempty"`
	Comments []Comment `json:"comments,omitempty"`
}

func (t *Task) IsOpen() bool {
	return !t.IsArchived && (t.Status == StatusNew || t.Status == StatusInProgress)
}

func (t *Task) IsClosed() bool {
	return t.IsArchived || t.Status == StatusDone || t.Status == StatusRejected
}

func (t *Task) CanManage(userID int64, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	return t.AssigneeID != 0 && t.AssigneeID == userID
}

type Subtask struct {
	ID        int64     `json:"id"`
	TaskID    string    `json:"task_id"`
	Title     string    `json:"title"`
	IsDone    bool      `json:"is_done"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID         int64     `json:"id"`
	TaskID     string    `json:"task_id"`
	AuthorID   int64     `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserSettings struct {
	UserID    int64     `json:"user_id"`
	NotifyDM  bool      `json:"notify_dm"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	UpdatedAt time.Time `json:"updated_at"`
}
