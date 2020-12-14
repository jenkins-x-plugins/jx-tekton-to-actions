package actions

// EventBase represents the base events
type EventBase struct {
	Types []string `json:"types,omitempty"`
}

// BranchEvent represents events with branches
type BranchEvent struct {
	EventBase
	Branches       []string `json:"branches,omitempty"`
	BranchesIgnore []string `json:"branches-ignore,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	TagsIgnore     []string `json:"tags-ignore,omitempty"`
	Paths          []string `json:"paths,omitempty"`
	PathsIgnore    []string `json:"paths-ignore,omitempty"`
}

// Events represents the events for a workflow
type Events struct {
	CheckRun           *EventBase   `json:"check_run,omitempty"`
	CheckSuite         *EventBase   `json:"check_suite,omitempty"`
	Create             *EventBase   `json:"create,omitempty"`
	Delete             *EventBase   `json:"delete,omitempty"`
	Deployment         *EventBase   `json:"deployment,omitempty"`
	DeploymentStatus   *EventBase   `json:"deployment_status,omitempty"`
	Fork               *EventBase   `json:"fork,omitempty"`
	Gollum             *EventBase   `json:"gollum,omitempty"`
	IssueComment       *EventBase   `json:"issue_comment,omitempty"`
	Issues             *EventBase   `json:"issues,omitempty"`
	Label              *EventBase   `json:"label,omitempty"`
	Milestone          *EventBase   `json:"milestone,omitempty"`
	PageBuild          *EventBase   `json:"page_build,omitempty"`
	Project            *EventBase   `json:"project,omitempty"`
	ProjectCard        *EventBase   `json:"project_card,omitempty"`
	ProjectColumn      *EventBase   `json:"project_column,omitempty"`
	Public             *EventBase   `json:"public,omitempty"`
	PullRequest        *BranchEvent `json:"pull_request,omitempty"`
	PullRequestReview  *EventBase   `json:"pull_request_review,omitempty"`
	PullRequestComment *EventBase   `json:"pull_request_comment,omitempty"`
	PullRequestTarget  *EventBase   `json:"pull_request_target,omitempty"`
	Push               *BranchEvent `json:"push,omitempty"`
	RegistryPackage    *EventBase   `json:"registry_package,omitempty"`
	Release            *EventBase   `json:"release,omitempty"`
	Status             *EventBase   `json:"status,omitempty"`
	Watch              *EventBase   `json:"watch,omitempty"`
	WorkflowRun        *EventBase   `json:"workflow_run,omitempty"`
}
