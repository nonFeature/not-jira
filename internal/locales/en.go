package locales

import (
	"fmt"
	"not-jira/internal/emoji"
)

func GetEn() *Bundle {
	return &Bundle{
		Common: Common{
			AdminOnly: fmt.Sprintf("%s Only administrators can create tasks.", emoji.Lock()),
			Cancelled: "Action cancelled.",
		},
		Start: Start{
			GreetingUser: fmt.Sprintf("%s <b>Hello! I am not-jira.</b>\n\n"+
				"I help collect and track tasks, bugs, and ideas directly in forum topics.\n\n"+
				"<b>Commands:</b>\n"+
				"• /list — View task list with filters and pages\n"+
				"• /view [ID] — View task card (e.g. /view B0)\n"+
				"• /settings — DM notification settings\n\n", emoji.Wave()),
			GreetingAdmin: fmt.Sprintf("%s <b>Administrator commands:</b>\n"+
				"• /add — Reply to a message in a forum topic to register a bug (B) or idea (I)\n"+
				"• In the task card you can change status, title, description, subtasks, and comments.\n", emoji.Star()),
		},
		Settings: Settings{
			Title:          fmt.Sprintf("%s <b>Notification Settings</b>\n\nDirect message notifications: <b>%%s</b>\n\nYou can change this setting using the button below:", emoji.Gear()),
			StatusEnabled:  fmt.Sprintf("enabled %s", emoji.Bell()),
			StatusDisabled: fmt.Sprintf("disabled %s", emoji.BellOff()),
			BtnEnable:      fmt.Sprintf("%s Enable DM notifications", emoji.Bell()),
			BtnDisable:     fmt.Sprintf("%s Disable DM notifications", emoji.BellOff()),
			UpdatedAlert:   "Settings updated",
		},
		Add: Add{
			UsageHint:          fmt.Sprintf("%s Reply with /add to a message with a bug/idea, or provide text: /add Bug description", emoji.Info()),
			ProcessingAI:       fmt.Sprintf("%s <i>Processing task with AI...</i>", emoji.Hourglass()),
			AcceptedBug:        fmt.Sprintf("%s Your bug <b>[%%s]</b> has been accepted!", emoji.Check()),
			AcceptedIdea:       fmt.Sprintf("%s Your idea <b>[%%s]</b> has been accepted!", emoji.Check()),
			CardHeader:         fmt.Sprintf("%s <b>Task Management:</b>\n\n", emoji.Clipboard()),
			FormTitlePrompt:    fmt.Sprintf("%s <b>Create task <code>[%%s]</code> (%%s)</b>\n\n<b>Original text:</b>\n<blockquote>%%s</blockquote>\n\nSend the <b>title</b> of the task (up to 60 characters):", emoji.Memo()),
			FormDescPrompt:     fmt.Sprintf("%s Title saved.\n\nNow enter the <b>description</b> (or send <code>-</code> to keep original text):", emoji.Check()),
			FormCreatedSuccess: fmt.Sprintf("%s <b>Task created successfully!</b>\n\n", emoji.Party()),
		},
		Edit: Edit{
			PromptEditTitle:    fmt.Sprintf("%s Enter new <b>title</b> for task <code>[%%s]</code>:", emoji.Pencil()),
			PromptEditDesc:     fmt.Sprintf("%s Enter new <b>description</b> for task <code>[%%s]</code>:", emoji.Memo()),
			PromptAddSubtask:   fmt.Sprintf("%s Enter name of <b>subtask</b> for <code>[%%s]</code>:", emoji.Plus()),
			PromptAddComment:   fmt.Sprintf("%s Enter text of <b>comment</b> for task <code>[%%s]</code>:", emoji.Messages()),
			TitleUpdated:       fmt.Sprintf("%s Task <b>[%%s]</b> title updated to: %%s", emoji.Check()),
			DescUpdated:        fmt.Sprintf("%s Task <b>[%%s]</b> description updated.", emoji.Check()),
			SubtaskAdded:       fmt.Sprintf("%s Subtask added to <b>[%%s]</b>: %%s", emoji.Check()),
			CommentAdded:       fmt.Sprintf("%s Comment added to <b>[%%s]</b>.", emoji.Check()),
			StatusChangedAlert: "Status changed: %s %s",
		},
		View: View{
			UsageHint:  fmt.Sprintf("%s Usage: /view B0 or /view I1", emoji.Info()),
			NotFound:   fmt.Sprintf("%s Task <code>%%s</code> not found.", emoji.Cross()),
			ListHeader: fmt.Sprintf("%s <b>Task List (%%d)</b>\nFilter: <code>%%s</code> | <code>%%s</code>\n\nSelect a task to view or edit:", emoji.Clipboard()),
			PageFormat: "Page %d/%d",
		},
		Task: Task{
			Header:         "<b>[%s]</b> %s <b>%s: %s</b>\n\n",
			StatusLabel:    "<b>Status:</b> %s %s\n",
			AuthorLabel:    "<b>Author:</b> @%s\n",
			CreatedLabel:   "<b>Created:</b> %s UTC\n\n",
			DescLabel:      "<b>Description:</b>\n<blockquote expandable>%s</blockquote>\n\n",
			SubtasksLabel:  "<b>Subtasks (%d/%d):</b>\n",
			CommentsLabel:  "<b>Comments (%d):</b>\n",
			TypeBug:        "Bug",
			TypeIdea:       "Idea",
			StatusNew:      "New",
			StatusProgress: "In Progress",
			StatusDone:     "Done",
			StatusRejected: "Rejected",
		},
		Buttons: Buttons{
			InProgress:   fmt.Sprintf("%s In Progress", emoji.Gear()),
			Done:         fmt.Sprintf("%s Done", emoji.Check()),
			Rejected:     fmt.Sprintf("%s Reject", emoji.Cross()),
			EditTitle:    fmt.Sprintf("%s Title", emoji.Pencil()),
			EditDesc:     fmt.Sprintf("%s Description", emoji.Memo()),
			AddSubtask:   fmt.Sprintf("%s Subtask", emoji.Plus()),
			AddComment:   fmt.Sprintf("%s Comment", emoji.Messages()),
			OriginalPost: fmt.Sprintf("%s Original Post", emoji.Link()),
			Refresh:      fmt.Sprintf("%s Refresh", emoji.Refresh()),
			PrevPage:     fmt.Sprintf("%s Back", emoji.ArrowL()),
			NextPage:     fmt.Sprintf("Next %s", emoji.ArrowR()),
			OpenDM:       fmt.Sprintf("%s Message Bot in DM", emoji.Rocket()),
		},
		Filters: Filters{
			AllTypes:    "All",
			Bugs:        fmt.Sprintf("%s Bugs", emoji.Bug()),
			Ideas:       fmt.Sprintf("%s Ideas", emoji.Idea()),
			AllStatuses: "All Statuses",
			New:         fmt.Sprintf("%s New", emoji.New()),
			InProgress:  fmt.Sprintf("%s In Progress", emoji.Gear()),
			Done:        fmt.Sprintf("%s Done", emoji.Check()),
		},
		Notifications: Notifications{
			PromptStartDM:      fmt.Sprintf("%s %%s, to allow the bot to message you, please start a conversation with it in DM first.", emoji.Warning()),
			TopicStatusUpdated: fmt.Sprintf("%s <b>Task status [%%s] updated!</b>\n\n<b>New status:</b> %%s %%s\n<b>Title:</b> %%s", emoji.Bell()),
			DMStatusUpdated:    fmt.Sprintf("%s <b>Update on your task [%%s]!</b>\n\n<b>New status:</b> %%s %%s\n<b>Title:</b> %%s\n\n<i>(You can disable DM notifications using the button below)</i>", emoji.Bell()),
		},
	}
}
