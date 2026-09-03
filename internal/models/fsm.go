package models

type FSMState string

const (
	StateNone              FSMState = ""
	StateCreatingTaskTitle FSMState = "creating_title"
	StateCreatingTaskDesc  FSMState = "creating_desc"
	StateEditingTitle      FSMState = "editing_title"
	StateEditingDesc       FSMState = "editing_desc"
	StateAddingSubtask     FSMState = "adding_subtask"
	StateAddingComment     FSMState = "adding_comment"
)

type UserSession struct {
	State     FSMState
	TaskID    string
	DraftTask *Task
}
