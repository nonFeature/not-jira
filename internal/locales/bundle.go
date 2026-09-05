package locales

type Bundle struct {
	Common        Common
	Start         Start
	Settings      Settings
	Add           Add
	Edit          Edit
	View          View
	Task          Task
	Buttons       Buttons
	Filters       Filters
	Notifications Notifications
}

type Common struct {
	AdminOnly string
	Cancelled string
}

type Start struct {
	GreetingUser  string
	GreetingAdmin string
}

type Settings struct {
	Title          string
	StatusEnabled  string
	StatusDisabled string
	BtnEnable      string
	BtnDisable     string
	UpdatedAlert   string
}

type Add struct {
	UsageHint          string
	ProcessingAI       string
	AcceptedBug        string
	AcceptedIdea       string
	CardHeader         string
	FormTitlePrompt    string
	FormDescPrompt     string
	FormCreatedSuccess string
}

type Edit struct {
	PromptEditTitle    string
	PromptEditDesc     string
	PromptAddSubtask   string
	PromptAddComment   string
	TitleUpdated       string
	DescUpdated        string
	SubtaskAdded           string
	CommentAdded           string
	StatusChangedAlert     string
	StatusReopenedAlert    string
	TaskArchivedAlert      string
	PromptTransfer         string
	TransferUserNotFound   string
	TransferOfferSent      string
	TransferOfferReceived  string
	TransferAcceptedNotify string
	TransferRejectedNotify string
	TaskClaimedNotify      string
}

type View struct {
	UsageHint  string
	NotFound   string
	ListHeader string
	PageFormat string
}

type Task struct {
	Header          string
	StatusLabel     string
	AuthorLabel     string
	AssigneeLabel   string
	UnassignedLabel string
	ArchivedBadge   string
	CreatedLabel    string
	DescLabel       string
	SubtasksLabel   string
	CommentsLabel   string
	TypeBug         string
	TypeIdea        string
	StatusNew       string
	StatusProgress  string
	StatusDone      string
	StatusRejected  string
}

type Buttons struct {
	InProgress   string
	Done         string
	Rejected     string
	Reopen       string
	Archive      string
	Claim        string
	Transfer     string
	AcceptTask   string
	RejectTask   string
	EditTitle    string
	EditDesc     string
	AddSubtask   string
	AddComment   string
	OriginalPost string
	Refresh      string
	PrevPage     string
	NextPage     string
	OpenDM       string
}

type Filters struct {
	AllTypes    string
	Bugs        string
	Ideas       string
	AllStatuses string
	New         string
	InProgress  string
	Done        string
}

type Notifications struct {
	PromptStartDM      string
	TopicStatusUpdated string
	DMStatusUpdated    string
}
