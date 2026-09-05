package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"not-jira/internal/models"
	"not-jira/internal/storage"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db *sql.DB
}

func New(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

func (s *SQLiteStorage) initSchema() error {
	createTables := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			num INTEGER NOT NULL,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			chat_id INTEGER NOT NULL,
			topic_id INTEGER NOT NULL,
			message_id INTEGER NOT NULL,
			message_link TEXT NOT NULL,
			author_id INTEGER NOT NULL,
			author_username TEXT NOT NULL,
			assignee_id INTEGER NOT NULL DEFAULT 0,
			assignee_username TEXT NOT NULL DEFAULT '',
			is_archived INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS subtasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			is_done INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			author_id INTEGER NOT NULL,
			author_name TEXT NOT NULL,
			text TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS user_settings (
			user_id INTEGER PRIMARY KEY,
			notify_dm INTEGER NOT NULL DEFAULT 1,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			username TEXT NOT NULL DEFAULT '',
			first_name TEXT NOT NULL DEFAULT '',
			updated_at DATETIME NOT NULL
		);`,
	}

	for _, q := range createTables {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}

	// Migrations for existing databases (must run before creating indexes on newly added columns)
	_, _ = s.db.Exec("ALTER TABLE tasks ADD COLUMN assignee_id INTEGER NOT NULL DEFAULT 0;")
	_, _ = s.db.Exec("ALTER TABLE tasks ADD COLUMN assignee_username TEXT NOT NULL DEFAULT '';")
	_, _ = s.db.Exec("ALTER TABLE tasks ADD COLUMN is_archived INTEGER NOT NULL DEFAULT 0;")

	createIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks(type);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_archived ON tasks(is_archived, status, updated_at);`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username COLLATE NOCASE);`,
	}

	for _, q := range createIndexes {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) GetNextTaskID(ctx context.Context, taskType models.TaskType) (string, int, error) {
	row := s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(num), -1) + 1 FROM tasks WHERE type = ?", taskType)
	var nextNum int
	if err := row.Scan(&nextNum); err != nil {
		return "", 0, fmt.Errorf("failed to calculate next task num: %w", err)
	}

	id := fmt.Sprintf("%s%d", taskType.Prefix(), nextNum)
	return id, nextNum, nil
}

func (s *SQLiteStorage) CreateTask(ctx context.Context, task *models.Task) error {
	now := time.Now().UTC()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.UpdatedAt = now

	query := `INSERT INTO tasks (
		id, num, type, title, description, status,
		chat_id, topic_id, message_id, message_link,
		author_id, author_username, assignee_id, assignee_username, is_archived,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	archivedInt := 0
	if task.IsArchived {
		archivedInt = 1
	}

	_, err := s.db.ExecContext(ctx, query,
		task.ID, task.Num, task.Type, task.Title, task.Description, task.Status,
		task.ChatID, task.TopicID, task.MessageID, task.MessageLink,
		task.AuthorID, task.AuthorUsername, task.AssigneeID, task.AssigneeUsername, archivedInt,
		task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (s *SQLiteStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	query := `SELECT
		id, num, type, title, description, status,
		chat_id, topic_id, message_id, message_link,
		author_id, author_username, assignee_id, assignee_username, is_archived,
		created_at, updated_at
	FROM tasks WHERE id = ? COLLATE NOCASE`

	row := s.db.QueryRowContext(ctx, query, id)
	var task models.Task
	var isArchivedInt int
	err := row.Scan(
		&task.ID, &task.Num, &task.Type, &task.Title, &task.Description, &task.Status,
		&task.ChatID, &task.TopicID, &task.MessageID, &task.MessageLink,
		&task.AuthorID, &task.AuthorUsername, &task.AssigneeID, &task.AssigneeUsername, &isArchivedInt,
		&task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	task.IsArchived = (isArchivedInt == 1)

	// Fetch subtasks
	subtasks, err := s.GetSubtasks(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	task.Subtasks = subtasks

	// Fetch comments
	comments, err := s.GetComments(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	task.Comments = comments

	return &task, nil
}

func (s *SQLiteStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now().UTC()
	archivedInt := 0
	if task.IsArchived {
		archivedInt = 1
	}

	query := `UPDATE tasks SET
		title = ?, description = ?, status = ?,
		assignee_id = ?, assignee_username = ?, is_archived = ?,
		updated_at = ?
	WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Status,
		task.AssigneeID, task.AssigneeUsername, archivedInt,
		task.UpdatedAt, task.ID,
	)
	return err
}

func (s *SQLiteStorage) DeleteTask(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) ListTasks(ctx context.Context, filter storage.TaskFilter, offset, limit int) ([]models.Task, int, error) {
	var whereClauses []string
	var args []interface{}

	if !filter.IncludeArchived {
		whereClauses = append(whereClauses, "is_archived = 0")
	}
	if filter.Type != nil {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, *filter.Type)
	}
	if filter.Status != nil {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.AssigneeID != nil {
		whereClauses = append(whereClauses, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM tasks" + whereSQL
	var totalCount int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Select page
	selectQuery := `SELECT
		id, num, type, title, description, status,
		chat_id, topic_id, message_id, message_link,
		author_id, author_username, assignee_id, assignee_username, is_archived,
		created_at, updated_at
	FROM tasks` + whereSQL + " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	pageArgs := append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, selectQuery, pageArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var isArchivedInt int
		if err := rows.Scan(
			&t.ID, &t.Num, &t.Type, &t.Title, &t.Description, &t.Status,
			&t.ChatID, &t.TopicID, &t.MessageID, &t.MessageLink,
			&t.AuthorID, &t.AuthorUsername, &t.AssigneeID, &t.AssigneeUsername, &isArchivedInt,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		t.IsArchived = (isArchivedInt == 1)
		tasks = append(tasks, t)
	}

	return tasks, totalCount, nil
}

func (s *SQLiteStorage) ArchiveInactiveClosedTasks(ctx context.Context, inactiveDuration time.Duration) (int64, error) {
	threshold := time.Now().UTC().Add(-inactiveDuration)
	query := `UPDATE tasks
		SET is_archived = 1
		WHERE is_archived = 0
		  AND status IN ('DONE', 'REJECTED')
		  AND updated_at <= ?`

	res, err := s.db.ExecContext(ctx, query, threshold)
	if err != nil {
		return 0, fmt.Errorf("failed to auto-archive inactive closed tasks: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s *SQLiteStorage) SetTaskUpdatedAt(ctx context.Context, id string, t time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE tasks SET updated_at = ? WHERE id = ?", t, id)
	return err
}

// Subtasks
func (s *SQLiteStorage) AddSubtask(ctx context.Context, taskID, title string) (*models.Subtask, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO subtasks (task_id, title, is_done, created_at) VALUES (?, ?, 0, ?)",
		taskID, title, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &models.Subtask{
		ID:        id,
		TaskID:    taskID,
		Title:     title,
		IsDone:    false,
		CreatedAt: now,
	}, nil
}

func (s *SQLiteStorage) ToggleSubtask(ctx context.Context, id int64) (*models.Subtask, error) {
	_, err := s.db.ExecContext(ctx, "UPDATE subtasks SET is_done = CASE WHEN is_done = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	var sub models.Subtask
	err = s.db.QueryRowContext(ctx, "SELECT id, task_id, title, is_done, created_at FROM subtasks WHERE id = ?", id).
		Scan(&sub.ID, &sub.TaskID, &sub.Title, &sub.IsDone, &sub.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SQLiteStorage) DeleteSubtask(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM subtasks WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) GetSubtasks(ctx context.Context, taskID string) ([]models.Subtask, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, task_id, title, is_done, created_at FROM subtasks WHERE task_id = ? ORDER BY id ASC", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subtasks []models.Subtask
	for rows.Next() {
		var sub models.Subtask
		if err := rows.Scan(&sub.ID, &sub.TaskID, &sub.Title, &sub.IsDone, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subtasks = append(subtasks, sub)
	}
	return subtasks, nil
}

// Comments
func (s *SQLiteStorage) AddComment(ctx context.Context, taskID string, authorID int64, authorName, text string) (*models.Comment, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO comments (task_id, author_id, author_name, text, created_at) VALUES (?, ?, ?, ?, ?)",
		taskID, authorID, authorName, text, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &models.Comment{
		ID:         id,
		TaskID:     taskID,
		AuthorID:   authorID,
		AuthorName: authorName,
		Text:       text,
		CreatedAt:  now,
	}, nil
}

func (s *SQLiteStorage) GetComments(ctx context.Context, taskID string) ([]models.Comment, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, task_id, author_id, author_name, text, created_at FROM comments WHERE task_id = ? ORDER BY created_at ASC", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.AuthorName, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// User Settings
func (s *SQLiteStorage) GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error) {
	var notify int
	var updated time.Time
	err := s.db.QueryRowContext(ctx, "SELECT notify_dm, updated_at FROM user_settings WHERE user_id = ?", userID).
		Scan(&notify, &updated)
	if err == sql.ErrNoRows {
		// Default: notifications enabled
		return &models.UserSettings{UserID: userID, NotifyDM: true}, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.UserSettings{UserID: userID, NotifyDM: notify == 1, UpdatedAt: updated}, nil
}

func (s *SQLiteStorage) SetNotifyDM(ctx context.Context, userID int64, notify bool) error {
	val := 0
	if notify {
		val = 1
	}
	now := time.Now().UTC()
	query := `INSERT INTO user_settings (user_id, notify_dm, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET notify_dm = excluded.notify_dm, updated_at = excluded.updated_at`
	_, err := s.db.ExecContext(ctx, query, userID, val, now)
	return err
}

// Users Cache
func (s *SQLiteStorage) UpsertUser(ctx context.Context, userID int64, username, firstName string) error {
	now := time.Now().UTC()
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	query := `INSERT INTO users (user_id, username, first_name, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			username = excluded.username,
			first_name = excluded.first_name,
			updated_at = excluded.updated_at`
	_, err := s.db.ExecContext(ctx, query, userID, username, firstName, now)
	return err
}

func (s *SQLiteStorage) FindUserIDByUsername(ctx context.Context, username string) (int64, error) {
	cleanUser := strings.TrimPrefix(strings.TrimSpace(username), "@")
	if cleanUser == "" {
		return 0, nil
	}
	query := `SELECT user_id FROM users WHERE username = ? COLLATE NOCASE LIMIT 1`
	var uid int64
	err := s.db.QueryRowContext(ctx, query, cleanUser).Scan(&uid)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return uid, nil
}
