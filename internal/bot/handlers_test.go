package bot

import (
	"strings"
	"testing"
	"time"

	"not-jira/internal/locales"
	"not-jira/internal/models"
)

func TestBuildPriorityPickerKeyboard(t *testing.T) {
	l := locales.ForUser("ru")
	kb := BuildPriorityPickerKeyboard("B12", models.PriorityP1, l)

	if kb == nil || len(kb.InlineKeyboard) < 3 {
		t.Fatalf("unexpected keyboard structure: %+v", kb)
	}

	// Row 0: P0, P1
	row0 := kb.InlineKeyboard[0]
	if len(row0) != 2 {
		t.Fatalf("expected 2 buttons in row 0, got %d", len(row0))
	}
	if row0[0].CallbackData != "set_priority:B12:P0" {
		t.Errorf("expected callback set_priority:B12:P0, got %s", row0[0].CallbackData)
	}
	if row0[1].CallbackData != "set_priority:B12:P1" {
		t.Errorf("expected callback set_priority:B12:P1, got %s", row0[1].CallbackData)
	}
	// P1 is current, should have markers
	if !strings.Contains(row0[1].Text, "•") {
		t.Errorf("expected marked text for P1, got %s", row0[1].Text)
	}
	if strings.Contains(row0[0].Text, "•") {
		t.Errorf("expected unmarked text for P0, got %s", row0[0].Text)
	}

	// Row 2: Back button
	row2 := kb.InlineKeyboard[2]
	if row2[0].CallbackData != "view:B12" {
		t.Errorf("expected callback view:B12 for back button, got %s", row2[0].CallbackData)
	}
}

func TestBuildMyTasksKeyboard(t *testing.T) {
	l := locales.ForUser("ru")
	tasks := []models.Task{
		{
			ID:     "B1",
			Title:  "Fix login crash",
			Status: models.StatusInProgress,
		},
	}

	kb := BuildMyTasksKeyboard(tasks, "assigned", 0, 2, l)
	if kb == nil || len(kb.InlineKeyboard) < 4 {
		t.Fatalf("unexpected keyboard structure: %+v", kb)
	}

	// Tab row
	tabRow := kb.InlineKeyboard[0]
	if !strings.Contains(tabRow[0].Text, "•") {
		t.Errorf("expected active tab mark on 'На мне', got %s", tabRow[0].Text)
	}
	if tabRow[0].CallbackData != "my:assigned:0" {
		t.Errorf("expected callback my:assigned:0, got %s", tabRow[0].CallbackData)
	}
	if tabRow[1].CallbackData != "my:created:0" {
		t.Errorf("expected callback my:created:0, got %s", tabRow[1].CallbackData)
	}

	// Task row
	taskRow := kb.InlineKeyboard[1]
	if taskRow[0].CallbackData != "view:B1" {
		t.Errorf("expected callback view:B1, got %s", taskRow[0].CallbackData)
	}

	// Navigation row
	navRow := kb.InlineKeyboard[2]
	// Page 0 of 2: should have page indicator and NextPage button
	hasNext := false
	for _, btn := range navRow {
		if btn.CallbackData == "my:assigned:1" {
			hasNext = true
		}
	}
	if !hasNext {
		t.Errorf("expected next page button with callback my:assigned:1")
	}

	// Bottom row: all tasks button
	bottomRow := kb.InlineKeyboard[3]
	if bottomRow[0].CallbackData != "list:ALL:ALL:ALL:0" {
		t.Errorf("expected callback list:ALL:ALL:ALL:0, got %s", bottomRow[0].CallbackData)
	}
}

func TestRenderTaskCardWithPriorityAndLabels(t *testing.T) {
	l := locales.ForUser("ru")
	task := &models.Task{
		ID:             "I5",
		Type:           models.TaskTypeIdea,
		Title:          "Dark mode support",
		Description:    "Please add a dark theme for nighttime usage.",
		Status:         models.StatusNew,
		Priority:       models.PriorityP1,
		Labels:         []string{"ui", "theme"},
		AuthorUsername: "tester",
		CreatedAt:      time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Now().UTC().Add(-2 * time.Minute),
		Subtasks: []models.Subtask{
			{Title: "Design palette", IsDone: true},
			{Title: "Implement CSS variables", IsDone: false},
		},
	}

	card := RenderTaskCard(task, l)

	if !strings.Contains(card, "Высокий") {
		t.Errorf("expected card to contain 'Высокий', got:\n%s", card)
	}
	if !strings.Contains(card, "🟡") {
		t.Errorf("expected card to contain '🟡', got:\n%s", card)
	}
	if !strings.Contains(card, "#ui #theme") {
		t.Errorf("expected card to contain '#ui #theme', got:\n%s", card)
	}
	if !strings.Contains(card, "<b>Теги:</b>") {
		t.Errorf("expected card to contain '<b>Теги:</b>', got:\n%s", card)
	}
	// Check subtasks strikethrough & progress bar
	if !strings.Contains(card, "<s>Design palette</s>") {
		t.Errorf("expected card to contain strikethrough '<s>Design palette</s>', got:\n%s", card)
	}
	if !strings.Contains(card, "▓") {
		t.Errorf("expected card to contain progress bar '▓', got:\n%s", card)
	}
	if !strings.Contains(card, "50%") {
		t.Errorf("expected card to contain 50%% progress, got:\n%s", card)
	}
	// Check relative time
	if !strings.Contains(card, "мин назад") {
		t.Errorf("expected card to contain relative updated time 'мин назад', got:\n%s", card)
	}
}

func TestTaskInlineKeyboardLabelsAndUnclaim(t *testing.T) {
	l := locales.ForUser("ru")
	task := &models.Task{
		ID:         "B42",
		Status:     models.StatusInProgress,
		AssigneeID: 1302340931,
	}

	kb := BuildTaskInlineKeyboard(task, 1302340931, true, false, l)
	if kb == nil {
		t.Fatalf("expected non-nil keyboard")
	}

	foundLabels := false
	foundUnclaim := false
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == "edit_labels:B42" {
				foundLabels = true
			}
			if btn.CallbackData == "unclaim:B42" {
				foundUnclaim = true
			}
		}
	}
	if !foundLabels {
		t.Errorf("expected button with callback edit_labels:B42 in task keyboard")
	}
	if !foundUnclaim {
		t.Errorf("expected button with callback unclaim:B42 in task keyboard for assignee")
	}
}

func TestBuildListKeyboardWithTag(t *testing.T) {
	l := locales.ForUser("ru")
	tasks := []models.Task{
		{ID: "B1", Title: "Task 1", Status: models.StatusNew},
	}

	kb := BuildListKeyboard(tasks, "BUG", "NEW", "frontend", 0, 1, l)
	if kb == nil {
		t.Fatalf("expected non-nil keyboard")
	}

	// Bottom row should have tag filter button
	bottomRow := kb.InlineKeyboard[len(kb.InlineKeyboard)-1]
	foundTagBtn := false
	for _, btn := range bottomRow {
		if btn.CallbackData == "list_tags:BUG:NEW:frontend:0" {
			foundTagBtn = true
			if !strings.Contains(btn.Text, "#frontend") {
				t.Errorf("expected tag button text to contain #frontend, got %s", btn.Text)
			}
		}
	}
	if !foundTagBtn {
		t.Errorf("expected list_tags:BUG:NEW:frontend:0 button in bottom row")
	}

	// Tag picker keyboard
	tagPicker := BuildListTagsKeyboard("BUG", "NEW", "frontend", 0, []string{"backend", "frontend"}, l)
	if tagPicker == nil {
		t.Fatalf("expected non-nil tagPicker keyboard")
	}
	foundFrontend := false
	for _, row := range tagPicker.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == "list:BUG:NEW:frontend:0" && btn.Text != l.Buttons.Back {
				foundFrontend = true
				if !strings.Contains(btn.Text, "• #frontend •") {
					t.Errorf("expected active marker on #frontend, got %s", btn.Text)
				}
			}
		}
	}
	if !foundFrontend {
		t.Errorf("expected list:BUG:NEW:frontend:0 button in tagPicker")
	}
}

func TestBuildTaskLabelsKeyboard(t *testing.T) {
	l := locales.ForUser("ru")
	task := &models.Task{
		ID:     "I0",
		Labels: []string{"litegram", "DotFL"},
	}

	available := []string{"litegram", "shortcuts", "DotFL", "other"}
	kb := BuildTaskLabelsKeyboard(task, available, l)
	if kb == nil {
		t.Fatalf("expected non-nil keyboard")
	}

	// Check that litegram is marked with checkmark ✅
	foundLitegram := false
	foundShortcuts := false
	foundManual := false
	foundClear := false
	foundBack := false

	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == "toggle_label:I0:litegram" {
				foundLitegram = true
				if !strings.Contains(btn.Text, "✅ #litegram") {
					t.Errorf("expected litegram button to have checkmark, got %s", btn.Text)
				}
			}
			if btn.CallbackData == "toggle_label:I0:shortcuts" {
				foundShortcuts = true
				if strings.Contains(btn.Text, "✅") {
					t.Errorf("expected shortcuts button not to have checkmark, got %s", btn.Text)
				}
			}
			if btn.CallbackData == "edit_labels_manual:I0" {
				foundManual = true
			}
			if btn.CallbackData == "clear_labels:I0" {
				foundClear = true
			}
			if btn.CallbackData == "view:I0" {
				foundBack = true
			}
		}
	}

	if !foundLitegram {
		t.Errorf("expected toggle_label:I0:litegram button")
	}
	if !foundShortcuts {
		t.Errorf("expected toggle_label:I0:shortcuts button")
	}
	if !foundManual {
		t.Errorf("expected edit_labels_manual:I0 button")
	}
	if !foundClear {
		t.Errorf("expected clear_labels:I0 button")
	}
	if !foundBack {
		t.Errorf("expected view:I0 back button")
	}
}

func TestBuildCancelKeyboard(t *testing.T) {
	l := locales.ForUser("ru")
	kbWithTask := BuildCancelKeyboard("B10", l)
	if kbWithTask == nil || len(kbWithTask.InlineKeyboard) != 1 {
		t.Fatalf("unexpected kbWithTask")
	}
	if kbWithTask.InlineKeyboard[0][0].CallbackData != "cancel_fsm:B10" {
		t.Errorf("expected callback cancel_fsm:B10, got %s", kbWithTask.InlineKeyboard[0][0].CallbackData)
	}

	kbEmpty := BuildCancelKeyboard("", l)
	if kbEmpty == nil || len(kbEmpty.InlineKeyboard) != 1 {
		t.Fatalf("unexpected kbEmpty")
	}
	if kbEmpty.InlineKeyboard[0][0].CallbackData != "cancel_fsm" {
		t.Errorf("expected callback cancel_fsm, got %s", kbEmpty.InlineKeyboard[0][0].CallbackData)
	}
}

func TestTaskLabelToggling(t *testing.T) {
	task := &models.Task{
		ID:     "B5",
		Labels: []string{"ui"},
	}

	if !task.HasLabel("ui") {
		t.Errorf("expected HasLabel('ui') = true")
	}
	if !task.HasLabel("#UI") {
		t.Errorf("expected HasLabel('#UI') = true with case and hash prefix")
	}
	if task.HasLabel("backend") {
		t.Errorf("expected HasLabel('backend') = false")
	}

	// Toggle backend on
	added := task.ToggleLabel("#backend")
	if !added || !task.HasLabel("backend") {
		t.Errorf("expected ToggleLabel to add backend")
	}

	// Toggle ui off
	added = task.ToggleLabel("UI")
	if added || task.HasLabel("ui") {
		t.Errorf("expected ToggleLabel to remove ui")
	}
}

func TestBuildSubtasksManageKeyboard(t *testing.T) {
	l := locales.ForUser("ru")
	task := &models.Task{
		ID: "B20",
		Subtasks: []models.Subtask{
			{ID: 1, TaskID: "B20", Title: "Subtask 1", IsDone: false},
			{ID: 2, TaskID: "B20", Title: "Subtask 2", IsDone: true},
		},
	}

	kb := BuildSubtasksManageKeyboard(task, l)
	if kb == nil || len(kb.InlineKeyboard) < 4 {
		t.Fatalf("unexpected keyboard structure: %+v", kb)
	}

	// Row 0: subtask 1 edit & delete
	row0 := kb.InlineKeyboard[0]
	if len(row0) != 2 {
		t.Fatalf("expected 2 buttons for subtask 1, got %d", len(row0))
	}
	if row0[0].CallbackData != "edit_sub:1:B20" {
		t.Errorf("expected callback edit_sub:1:B20, got %s", row0[0].CallbackData)
	}
	if !strings.Contains(row0[0].Text, "Subtask 1") {
		t.Errorf("expected text to contain 'Subtask 1', got %s", row0[0].Text)
	}
	if row0[1].CallbackData != "del_sub:1:B20" {
		t.Errorf("expected callback del_sub:1:B20, got %s", row0[1].CallbackData)
	}

	// Row 1: subtask 2 edit & delete
	row1 := kb.InlineKeyboard[1]
	if row1[0].CallbackData != "edit_sub:2:B20" {
		t.Errorf("expected callback edit_sub:2:B20, got %s", row1[0].CallbackData)
	}
	if row1[1].CallbackData != "del_sub:2:B20" {
		t.Errorf("expected callback del_sub:2:B20, got %s", row1[1].CallbackData)
	}

	// Action row: Add & Clear all
	actionRow := kb.InlineKeyboard[2]
	if len(actionRow) != 2 {
		t.Fatalf("expected 2 buttons in action row, got %d", len(actionRow))
	}
	if actionRow[0].CallbackData != "add_sub_prompt:B20" {
		t.Errorf("expected callback add_sub_prompt:B20, got %s", actionRow[0].CallbackData)
	}
	if actionRow[1].CallbackData != "clear_subs:B20" {
		t.Errorf("expected callback clear_subs:B20, got %s", actionRow[1].CallbackData)
	}

	// Back row
	backRow := kb.InlineKeyboard[3]
	if backRow[0].CallbackData != "view:B20" {
		t.Errorf("expected callback view:B20 for back button, got %s", backRow[0].CallbackData)
	}
}

func TestBuildCommentsManageKeyboard(t *testing.T) {
	l := locales.ForUser("ru")
	task := &models.Task{
		ID: "B21",
		Comments: []models.Comment{
			{ID: 10, TaskID: "B21", AuthorID: 100, AuthorName: "user1", Text: "First comment"},
			{ID: 11, TaskID: "B21", AuthorID: 200, AuthorName: "user2", Text: "Second comment"},
		},
	}

	// User 100 (non-admin) viewing: can edit/delete comment 10, but not 11
	kbUser := BuildCommentsManageKeyboard(task, 100, false, l)
	if kbUser == nil || len(kbUser.InlineKeyboard) < 4 {
		t.Fatalf("unexpected kbUser structure")
	}
	// Comment 10: editable
	if len(kbUser.InlineKeyboard[0]) != 2 || kbUser.InlineKeyboard[0][0].CallbackData != "edit_comm:10:B21" {
		t.Errorf("expected editable comment 10 for author")
	}
	// Comment 11: read-only for user 100
	if len(kbUser.InlineKeyboard[1]) != 1 || kbUser.InlineKeyboard[1][0].CallbackData != "noop" {
		t.Errorf("expected read-only comment 11 for non-author")
	}
	// Action row: only Add button (no Clear All for non-admin)
	if len(kbUser.InlineKeyboard[2]) != 1 || kbUser.InlineKeyboard[2][0].CallbackData != "add_comm_prompt:B21" {
		t.Errorf("expected only Add button in action row for non-admin")
	}

	// Admin viewing: can edit/delete all and clear all
	kbAdmin := BuildCommentsManageKeyboard(task, 999, true, l)
	if len(kbAdmin.InlineKeyboard[0]) != 2 || kbAdmin.InlineKeyboard[0][0].CallbackData != "edit_comm:10:B21" {
		t.Errorf("expected editable comment 10 for admin")
	}
	if len(kbAdmin.InlineKeyboard[1]) != 2 || kbAdmin.InlineKeyboard[1][0].CallbackData != "edit_comm:11:B21" {
		t.Errorf("expected editable comment 11 for admin")
	}
	if len(kbAdmin.InlineKeyboard[2]) != 2 || kbAdmin.InlineKeyboard[2][1].CallbackData != "clear_comms:B21" {
		t.Errorf("expected Clear all comments button for admin")
	}
}

func TestTaskInlineKeyboardManageSubAndComm(t *testing.T) {
	l := locales.ForUser("ru")
	task := &models.Task{
		ID:         "B30",
		Status:     models.StatusInProgress,
		AssigneeID: 1302340931,
	}

	kb := BuildTaskInlineKeyboard(task, 1302340931, true, false, l)
	foundManageSub := false
	foundManageComm := false
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == "manage_sub:B30" {
				foundManageSub = true
			}
			if btn.CallbackData == "manage_comm:B30" {
				foundManageComm = true
			}
		}
	}
	if !foundManageSub {
		t.Errorf("expected manage_sub:B30 button in task inline keyboard")
	}
	if !foundManageComm {
		t.Errorf("expected manage_comm:B30 button in task inline keyboard")
	}
}

