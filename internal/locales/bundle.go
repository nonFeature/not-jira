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
	Backup        Backup
}

type Backup struct {
	Creating string
	Caption  string
	Failed   string
}

type Common struct {
	AdminOnly       string
	Cancelled       string
	OnlyTextAllowed string
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
	PromptAddSubtask       string
	PromptEditSubtask      string
	PromptAddComment       string
	PromptEditComment      string
	TitleUpdated           string
	DescUpdated            string
	SubtaskAdded           string
	SubtaskUpdated         string
	SubtaskDeletedAlert    string
	SubtasksClearedAlert   string
	CommentAdded           string
	CommentUpdated         string
	CommentDeletedAlert    string
	CommentsClearedAlert   string
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
	PromptEditLabels       string
	LabelsUpdated          string
	LabelToggledOnAlert    string
	LabelToggledOffAlert   string
	LabelsClearedAlert     string
	PriorityChangedAlert   string
	TaskUnclaimedAlert     string
	CannotEditOtherComment string
}

type View struct {
	UsageHint     string
	NotFound      string
	ListHeader    string
	MyTasksHeader string
	NoMyTasks     string
	PageFormat    string
}

type Task struct {
	Header          string
	StatusLabel     string
	PriorityLabel   string
	LabelsLabel     string
	AuthorLabel     string
	AssigneeLabel   string
	UnassignedLabel string
	ArchivedBadge   string
	CreatedLabel    string
	UpdatedLabel    string
	JustNow         string
	MinutesAgo      string
	HoursAgo        string
	DaysAgo         string
	DescLabel       string
	SubtasksLabel   string
	CommentsLabel   string
	TypeBug         string
	TypeIdea        string
	StatusNew       string
	StatusProgress  string
	StatusDone      string
	StatusRejected  string
	P0              string
	P1              string
	P2              string
	P3              string
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
	Priority     string
	Labels       string
	MyTasks      string
	EditTitle    string
	EditDesc     string
	AddSubtask   string
	AddComment   string
	OriginalPost string
	Refresh      string
	PrevPage         string
	NextPage         string
	OpenDM           string
	EditLabelsManual string
	ClearLabels      string
	Back             string
	Unclaim          string
	TagFilter        string
	Cancel           string
	BackToTask       string
	ClearAll         string
	Add              string
	EditAction       string
	DeleteAction     string
	BackToSubtasks   string
	BackToComments   string
}

type Filters struct {
	AllTypes    string
	Bugs        string
	Ideas       string
	AllStatuses string
	New         string
	InProgress  string
	Done        string
	MyAssigned  string
	MyCreated   string
	AllTags     string
}

type Notifications struct {
	PromptStartDM      string
	TopicStatusUpdated string
	DMStatusUpdated    string
}
