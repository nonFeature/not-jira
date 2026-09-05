package storage

import (
	"context"
	"time"

	"not-jira/internal/models"
)

type TaskFilter struct {
	Type            *models.TaskType
	Status          *models.TaskStatus
	IncludeArchived bool
	AssigneeID      *int64
}

type Storage interface {
	// Tasks
	GetNextTaskID(ctx context.Context, taskType models.TaskType) (id string, num int, err error)
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filter TaskFilter, offset, limit int) ([]models.Task, int, error)
	ArchiveInactiveClosedTasks(ctx context.Context, inactiveDuration time.Duration) (int64, error)

	// Subtasks
	AddSubtask(ctx context.Context, taskID, title string) (*models.Subtask, error)
	ToggleSubtask(ctx context.Context, id int64) (*models.Subtask, error)
	DeleteSubtask(ctx context.Context, id int64) error
	GetSubtasks(ctx context.Context, taskID string) ([]models.Subtask, error)

	// Comments
	AddComment(ctx context.Context, taskID string, authorID int64, authorName, text string) (*models.Comment, error)
	GetComments(ctx context.Context, taskID string) ([]models.Comment, error)

	// User Settings
	GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error)
	SetNotifyDM(ctx context.Context, userID int64, notify bool) error

	// Users Cache
	UpsertUser(ctx context.Context, userID int64, username, firstName string) error
	FindUserIDByUsername(ctx context.Context, username string) (int64, error)

	Close() error
}
