package sqlite_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"not-jira/internal/models"
	"not-jira/internal/storage"
	"not-jira/internal/storage/sqlite"
)

func setupTestDB(t *testing.T) (*sqlite.SQLiteStorage, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_bot.db")

	st, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite: %v", err)
	}

	cleanup := func() {
		_ = st.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return st, cleanup
}

func TestTaskIDGeneration(t *testing.T) {
	st, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// First bug should be B0
	idB0, numB0, err := st.GetNextTaskID(ctx, models.TaskTypeBug)
	if err != nil {
		t.Fatalf("GetNextTaskID bug failed: %v", err)
	}
	if idB0 != "B0" || numB0 != 0 {
		t.Errorf("expected B0 (num 0), got %s (num %d)", idB0, numB0)
	}

	// Insert B0
	err = st.CreateTask(ctx, &models.Task{
		ID:          idB0,
		Num:         numB0,
		Type:        models.TaskTypeBug,
		Title:       "Test Bug 0",
		Description: "Desc",
		Status:      models.StatusNew,
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Next bug should be B1
	idB1, numB1, err := st.GetNextTaskID(ctx, models.TaskTypeBug)
	if err != nil {
		t.Fatalf("GetNextTaskID bug 1 failed: %v", err)
	}
	if idB1 != "B1" || numB1 != 1 {
		t.Errorf("expected B1 (num 1), got %s (num %d)", idB1, numB1)
	}

	// First idea should be I0
	idI0, numI0, err := st.GetNextTaskID(ctx, models.TaskTypeIdea)
	if err != nil {
		t.Fatalf("GetNextTaskID idea failed: %v", err)
	}
	if idI0 != "I0" || numI0 != 0 {
		t.Errorf("expected I0 (num 0), got %s (num %d)", idI0, numI0)
	}
}

func TestSubtasksAndComments(t *testing.T) {
	st, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	task := &models.Task{
		ID:          "B0",
		Num:         0,
		Type:        models.TaskTypeBug,
		Title:       "Network issue",
		Description: "WiFi drops when VPN runs",
		Status:      models.StatusNew,
		ChatID:      -10012345,
	}

	if err := st.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Add subtasks
	sub1, err := st.AddSubtask(ctx, "B0", "Check MT7981 PPE offload")
	if err != nil {
		t.Fatalf("AddSubtask failed: %v", err)
	}
	if sub1.IsDone {
		t.Errorf("expected new subtask to be not done")
	}

	// Toggle subtask
	toggled, err := st.ToggleSubtask(ctx, sub1.ID)
	if err != nil {
		t.Fatalf("ToggleSubtask failed: %v", err)
	}
	if !toggled.IsDone {
		t.Errorf("expected toggled subtask to be done")
	}

	// Add comment
	comm, err := st.AddComment(ctx, "B0", 12345, "dev_admin", "Confirmed PPE conflict with WED")
	if err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
	if comm.Text != "Confirmed PPE conflict with WED" {
		t.Errorf("unexpected comment text: %s", comm.Text)
	}

	// Fetch task and check relations
	loaded, err := st.GetTask(ctx, "B0")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if len(loaded.Subtasks) != 1 {
		t.Errorf("expected 1 subtask, got %d", len(loaded.Subtasks))
	}
	if len(loaded.Comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(loaded.Comments))
	}

	// Update subtask
	if err := st.UpdateSubtask(ctx, sub1.ID, "Updated subtask title"); err != nil {
		t.Fatalf("UpdateSubtask failed: %v", err)
	}
	subs, _ := st.GetSubtasks(ctx, "B0")
	if len(subs) != 1 || subs[0].Title != "Updated subtask title" {
		t.Errorf("expected updated subtask title, got %v", subs)
	}

	// Update comment
	if err := st.UpdateComment(ctx, comm.ID, "Updated comment text"); err != nil {
		t.Fatalf("UpdateComment failed: %v", err)
	}
	comms, _ := st.GetComments(ctx, "B0")
	if len(comms) != 1 || comms[0].Text != "Updated comment text" {
		t.Errorf("expected updated comment text, got %v", comms)
	}

	// Delete subtask
	if err := st.DeleteSubtask(ctx, sub1.ID); err != nil {
		t.Fatalf("DeleteSubtask failed: %v", err)
	}
	subsAfterDel, _ := st.GetSubtasks(ctx, "B0")
	if len(subsAfterDel) != 0 {
		t.Errorf("expected 0 subtasks after delete, got %d", len(subsAfterDel))
	}

	// Delete comment
	if err := st.DeleteComment(ctx, comm.ID); err != nil {
		t.Fatalf("DeleteComment failed: %v", err)
	}
	commsAfterDel, _ := st.GetComments(ctx, "B0")
	if len(commsAfterDel) != 0 {
		t.Errorf("expected 0 comments after delete, got %d", len(commsAfterDel))
	}

	// Test ClearSubtasks and ClearComments
	_, _ = st.AddSubtask(ctx, "B0", "Sub 1")
	_, _ = st.AddSubtask(ctx, "B0", "Sub 2")
	_ = st.ClearSubtasks(ctx, "B0")
	subsAfterClear, _ := st.GetSubtasks(ctx, "B0")
	if len(subsAfterClear) != 0 {
		t.Errorf("expected 0 subtasks after clear, got %d", len(subsAfterClear))
	}

	_, _ = st.AddComment(ctx, "B0", 1, "test", "Comm 1")
	_ = st.ClearComments(ctx, "B0")
	commsAfterClear, _ := st.GetComments(ctx, "B0")
	if len(commsAfterClear) != 0 {
		t.Errorf("expected 0 comments after clear, got %d", len(commsAfterClear))
	}
}

func TestListAndPagination(t *testing.T) {
	st, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Insert 7 bugs
	for i := 0; i < 7; i++ {
		id, num, _ := st.GetNextTaskID(ctx, models.TaskTypeBug)
		_ = st.CreateTask(ctx, &models.Task{
			ID:          id,
			Num:         num,
			Type:        models.TaskTypeBug,
			Title:       id,
			Description: "desc",
			Status:      models.StatusNew,
		})
	}

	// Page 0 (limit 5)
	tasksP0, total, err := st.ListTasks(ctx, storage.TaskFilter{}, 0, 5)
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if total != 7 {
		t.Errorf("expected total 7, got %d", total)
	}
	if len(tasksP0) != 5 {
		t.Errorf("expected 5 tasks on page 0, got %d", len(tasksP0))
	}

	// Page 1 (limit 5, offset 5)
	tasksP1, _, err := st.ListTasks(ctx, storage.TaskFilter{}, 5, 5)
	if err != nil {
		t.Fatalf("ListTasks p1 failed: %v", err)
	}
	if len(tasksP1) != 2 {
		t.Errorf("expected 2 tasks on page 1, got %d", len(tasksP1))
	}
}

func TestUserSettings(t *testing.T) {
	st, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID := int64(987654)

	// Default should be notify = true
	settings, err := st.GetUserSettings(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserSettings failed: %v", err)
	}
	if !settings.NotifyDM {
		t.Errorf("expected default NotifyDM to be true")
	}

	// Disable DM
	err = st.SetNotifyDM(ctx, userID, false)
	if err != nil {
		t.Fatalf("SetNotifyDM failed: %v", err)
	}

	updated, err := st.GetUserSettings(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserSettings updated failed: %v", err)
	}
	if updated.NotifyDM {
		t.Errorf("expected NotifyDM to be false")
	}
}

func TestAutoArchiveAndUsers(t *testing.T) {
	st, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test Users
	err := st.UpsertUser(ctx, 12345, "johndoe", "John")
	if err != nil {
		t.Fatalf("UpsertUser failed: %v", err)
	}
	uid, err := st.FindUserIDByUsername(ctx, "@johndoe")
	if err != nil || uid != 12345 {
		t.Errorf("expected user ID 12345, got %d, err: %v", uid, err)
	}

	// Test Task with Assignee
	task := &models.Task{
		ID:               "B0",
		Num:              0,
		Type:             models.TaskTypeBug,
		Title:            "Old bug",
		Description:      "Desc",
		Status:           models.StatusDone,
		AssigneeID:       12345,
		AssigneeUsername: "johndoe",
	}
	if err := st.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Active task that is new
	taskNew := &models.Task{
		ID:          "B1",
		Num:         1,
		Type:        models.TaskTypeBug,
		Title:       "New bug",
		Description: "Desc",
		Status:      models.StatusNew,
	}
	if err := st.CreateTask(ctx, taskNew); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Verify task retrieval with assignee
	retrieved, err := st.GetTask(ctx, "B0")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if retrieved.AssigneeID != 12345 || retrieved.AssigneeUsername != "johndoe" {
		t.Errorf("expected assignee 12345/johndoe, got %d/%s", retrieved.AssigneeID, retrieved.AssigneeUsername)
	}

	// Manually set updated_at of B0 to 10 days ago (more than 7 days)
	oldTime := time.Now().UTC().Add(-10 * 24 * time.Hour)
	if err := st.SetTaskUpdatedAt(ctx, task.ID, oldTime); err != nil {
		t.Fatalf("SetTaskUpdatedAt failed: %v", err)
	}

	// Run auto-archive with 7 days threshold
	archivedCount, err := st.ArchiveInactiveClosedTasks(ctx, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("ArchiveInactiveClosedTasks failed: %v", err)
	}
	if archivedCount != 1 {
		t.Errorf("expected 1 task archived, got %d", archivedCount)
	}

	// Verify B0 is archived and B1 is not
	b0, _ := st.GetTask(ctx, "B0")
	if !b0.IsArchived {
		t.Errorf("expected B0 to be archived")
	}
	b1, _ := st.GetTask(ctx, "B1")
	if b1.IsArchived {
		t.Errorf("expected B1 not to be archived")
	}

	// Verify ListTasks without IncludeArchived only returns B1
	list, total, err := st.ListTasks(ctx, storage.TaskFilter{}, 0, 10)
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].ID != "B1" {
		t.Errorf("expected ListTasks to only return 1 active task B1, got total %d", total)
	}
}

func TestLegacyMigration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "not-jira-mig-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "legacy.db")

	// 1. Create a legacy database without is_archived and assignee columns
	rawDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open raw sqlite: %v", err)
	}
	legacySchema := `
		CREATE TABLE tasks (
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
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`
	if _, err := rawDB.Exec(legacySchema); err != nil {
		rawDB.Close()
		t.Fatalf("failed to create legacy schema: %v", err)
	}
	rawDB.Close()

	// 2. Open via sqlite.New(dbPath) which runs initSchema and migrations
	st, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("sqlite.New(dbPath) failed on legacy database: %v", err)
	}
	defer st.Close()

	// 3. Verify that new columns and tables work
	ctx := context.Background()
	task := &models.Task{
		ID:               "L0",
		Num:              0,
		Type:             models.TaskTypeBug,
		Title:            "Legacy task",
		Description:      "Testing migration",
		Status:           models.StatusNew,
		AssigneeID:       999,
		AssigneeUsername: "migrated_user",
		IsArchived:       false,
	}
	if err := st.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed after migration: %v", err)
	}

	got, err := st.GetTask(ctx, "L0")
	if err != nil {
		t.Fatalf("GetTask failed after migration: %v", err)
	}
	if got.AssigneeID != 999 || got.AssigneeUsername != "migrated_user" {
		t.Errorf("expected assignee 999/migrated_user, got %d/%s", got.AssigneeID, got.AssigneeUsername)
	}
}

func TestPriorityAndLabels(t *testing.T) {
	st, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := &models.Task{
		ID:          "P0_TASK",
		Num:         100,
		Type:        models.TaskTypeBug,
		Title:       "Critical Auth Bug",
		Description: "Login fails",
		Status:      models.StatusNew,
		Priority:    models.PriorityP0,
		Labels:      []string{"backend", "auth"},
		AuthorID:    444,
	}

	if err := st.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	got, err := st.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Priority != models.PriorityP0 {
		t.Errorf("expected priority P0, got %s", got.Priority)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "backend" || got.Labels[1] != "auth" {
		t.Errorf("expected labels [backend, auth], got %v", got.Labels)
	}

	// Update priority and labels
	got.Priority = models.PriorityP1
	got.Labels = []string{"security"}
	if err := st.UpdateTask(ctx, got); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	updated, _ := st.GetTask(ctx, task.ID)
	if updated.Priority != models.PriorityP1 {
		t.Errorf("expected updated priority P1, got %s", updated.Priority)
	}
	if len(updated.Labels) != 1 || updated.Labels[0] != "security" {
		t.Errorf("expected updated labels [security], got %v", updated.Labels)
	}

	// Filter by Label
	tasksByLabel, total, err := st.ListTasks(ctx, storage.TaskFilter{Label: "security"}, 0, 10)
	if err != nil || total != 1 || len(tasksByLabel) != 1 {
		t.Errorf("expected 1 task by label 'security', got total %d, err: %v", total, err)
	}

	// Filter by Author
	authID := int64(444)
	tasksByAuthor, total, err := st.ListTasks(ctx, storage.TaskFilter{AuthorID: &authID}, 0, 10)
	if err != nil || total != 1 || len(tasksByAuthor) != 1 {
		t.Errorf("expected 1 task by AuthorID 444, got total %d, err: %v", total, err)
	}

	// Test GetAllLabels
	labels, err := st.GetAllLabels(ctx)
	if err != nil {
		t.Fatalf("GetAllLabels failed: %v", err)
	}
	if len(labels) != 1 || labels[0] != "security" {
		t.Errorf("expected labels ['security'], got %v", labels)
	}
}


