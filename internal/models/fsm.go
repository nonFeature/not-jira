package models

type FSMState string

const (
	StateNone              FSMState = ""
	StateCreatingTaskTitle FSMState = "creating_title"
	StateCreatingTaskDesc  FSMState = "creating_desc"
	StateEditingTitle      FSMState = "editing_title"
	StateEditingDesc       FSMState = "editing_desc"
	StateAddingSubtask     FSMState = "adding_subtask"
	StateEditingSubtask    FSMState = "editing_subtask"
	StateAddingComment     FSMState = "adding_comment"
	StateEditingComment    FSMState = "editing_comment"
	StateAssigningTask     FSMState = "assigning_task"
	StateEditingLabels     FSMState = "editing_labels"
)

type UserSession struct {
	State     FSMState
	TaskID    string
	SubtaskID int64
	CommentID int64
	DraftTask *Task
}
