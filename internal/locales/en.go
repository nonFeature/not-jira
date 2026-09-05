package locales

import (
	"fmt"
	"not-jira/internal/emoji"
)

func GetEn() *Bundle {
	return &Bundle{
		Common: Common{
			AdminOnly: fmt.Sprintf("%s Admin only.", emoji.Lock()),
			Cancelled: "Cancelled.",
		},
		Start: Start{
			GreetingUser: fmt.Sprintf("%s <b>not-jira</b> — topic task tracker\n\n"+
				"<b>Commands:</b>\n"+
				"/list — task list\n"+
				"/view [ID] — task card (e.g. /view B0)\n"+
				"/settings — DM notifications\n\n", emoji.Wave()),
			GreetingAdmin: fmt.Sprintf("%s <b>Admins:</b>\n"+
				"/add — create task (reply to message)\n", emoji.Star()),
		},
		Settings: Settings{
			Title:          fmt.Sprintf("%s <b>Notification Settings</b>\n\nDM notifications: <b>%%s</b>", emoji.Gear()),
			StatusEnabled:  fmt.Sprintf("enabled %s", emoji.Bell()),
			StatusDisabled: fmt.Sprintf("disabled %s", emoji.BellOff()),
			BtnEnable:      "Enable",
			BtnDisable:     "Disable",
			UpdatedAlert:   "Settings updated",
		},
		Add: Add{
			UsageHint:          fmt.Sprintf("%s Reply with /add to a message or provide text: /add [description]", emoji.Info()),
			ProcessingAI:       fmt.Sprintf("%s <i>Processing...</i>", emoji.Hourglass()),
			AcceptedBug:        fmt.Sprintf("%s Bug <b>[%%s]</b> created", emoji.Check()),
			AcceptedIdea:       fmt.Sprintf("%s Idea <b>[%%s]</b> created", emoji.Check()),
			CardHeader:         fmt.Sprintf("%s <b>Task Card:</b>\n\n", emoji.Clipboard()),
			FormTitlePrompt:    fmt.Sprintf("%s <b>New task <code>[%%s]</code> (%%s)</b>\n\n<blockquote>%%s</blockquote>\n\nSend <b>title</b> (up to 60 chars):", emoji.Memo()),
			FormDescPrompt:     fmt.Sprintf("%s Title saved.\n\nEnter <b>description</b> (or <code>-</code> to keep original text):", emoji.Check()),
			FormCreatedSuccess: fmt.Sprintf("%s <b>Task created!</b>\n\n", emoji.Party()),
		},
		Edit: Edit{
			PromptEditTitle:        fmt.Sprintf("%s Enter <b>title</b> for <code>[%%s]</code>:", emoji.Pencil()),
			PromptEditDesc:         fmt.Sprintf("%s Enter <b>description</b> for <code>[%%s]</code>:", emoji.Memo()),
			PromptAddSubtask:       fmt.Sprintf("%s Enter <b>subtask</b> for <code>[%%s]</code>:", emoji.Plus()),
			PromptAddComment:       fmt.Sprintf("%s Enter <b>comment</b> for <code>[%%s]</code>:", emoji.Messages()),
			TitleUpdated:           fmt.Sprintf("%s Title <b>[%%s]</b> updated: %%s", emoji.Check()),
			DescUpdated:            fmt.Sprintf("%s Description <b>[%%s]</b> updated.", emoji.Check()),
			SubtaskAdded:           fmt.Sprintf("%s Subtask added to <b>[%%s]</b>: %%s", emoji.Check()),
			CommentAdded:           fmt.Sprintf("%s Comment added to <b>[%%s]</b>.", emoji.Check()),
			StatusChangedAlert:     "Status: %s %s",
			StatusReopenedAlert:    "Task reopened",
			TaskArchivedAlert:      "Task moved to archive",
			PromptTransfer:         fmt.Sprintf("%s Transfer task <code>[%%s]</code> to?\nEnter @username or forward user's message:", emoji.User()),
			TransferUserNotFound:   "User @%s not found. Ask them to start the bot in DM first.",
			TransferOfferSent:      "Transfer offer sent to @%s in DM.",
			TransferOfferReceived:  fmt.Sprintf("%s @%%s offers you task <b>[%%s]</b>:\n\n<b>%%s</b>\n<blockquote>%%s</blockquote>\n\nDo you want to accept this task?", emoji.User()),
			TransferAcceptedNotify: fmt.Sprintf("%s @%%s accepted task <b>[%%s]</b>.", emoji.Check()),
			TransferRejectedNotify: fmt.Sprintf("%s @%%s declined task <b>[%%s]</b>.", emoji.Cross()),
			TaskClaimedNotify:      fmt.Sprintf("%s @%%s claimed task <b>[%%s]</b>.", emoji.Check()),
		},
		View: View{
			UsageHint:  fmt.Sprintf("%s Example: /view B0", emoji.Info()),
			NotFound:   fmt.Sprintf("%s Task <code>%%s</code> not found.", emoji.Cross()),
			ListHeader: fmt.Sprintf("%s <b>Tasks (%%d)</b>\nFilter: <code>%%s</code> | <code>%%s</code>\n\n", emoji.Clipboard()),
			PageFormat: "Page %d/%d",
		},
		Task: Task{
			Header:          "<b>[%s]</b> %s <b>%s: %s</b>\n\n",
			StatusLabel:     "<b>Status:</b> %s %s\n",
			AuthorLabel:     "<b>Author:</b> @%s\n",
			AssigneeLabel:   "<b>Assignee:</b> @%s\n",
			UnassignedLabel: "<b>Assignee:</b> <i>unassigned</i>\n",
			ArchivedBadge:   fmt.Sprintf(" %s <i>(archived)</i>", emoji.Box()),
			CreatedLabel:    "<b>Created:</b> %s UTC\n\n",
			DescLabel:       "<b>Description:</b>\n<blockquote expandable>%s</blockquote>\n\n",
			SubtasksLabel:   "<b>Subtasks (%d/%d):</b>\n",
			CommentsLabel:   "<b>Comments (%d):</b>\n",
			TypeBug:         "Bug",
			TypeIdea:        "Idea",
			StatusNew:       "New",
			StatusProgress:  "In Progress",
			StatusDone:      "Done",
			StatusRejected:  "Rejected",
		},
		Buttons: Buttons{
			InProgress:   "In Progress",
			Done:         "Done",
			Rejected:     "Reject",
			Reopen:       "Reopen",
			Archive:      "Archive",
			Claim:        "Claim",
			Transfer:     "Transfer",
			AcceptTask:   "Accept Task",
			RejectTask:   "Decline",
			EditTitle:    "Title",
			EditDesc:     "Description",
			AddSubtask:   "Subtask",
			AddComment:   "Comment",
			OriginalPost: "Original Post",
			Refresh:      "Refresh",
			PrevPage:     "Back",
			NextPage:     "Next",
			OpenDM:       "Open DM",
		},
		Filters: Filters{
			AllTypes:    "All",
			Bugs:        "Bugs",
			Ideas:       "Ideas",
			AllStatuses: "All Statuses",
			New:         "New",
			InProgress:  "In Progress",
			Done:        "Done",
		},
		Notifications: Notifications{
			PromptStartDM:      fmt.Sprintf("%s %%s, please message the bot in DM first.", emoji.Warning()),
			TopicStatusUpdated: fmt.Sprintf("%s <b>[%%s]</b> %%s %%s\n%%s", emoji.Bell()),
			DMStatusUpdated:    fmt.Sprintf("%s <b>[%%s]</b> %%s %%s\n%%s", emoji.Bell()),
		},
	}
}
