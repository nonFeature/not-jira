package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
