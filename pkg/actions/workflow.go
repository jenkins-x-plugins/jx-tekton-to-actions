package actions

// Workflow represents a github workflow
type Workflow struct {
	Name string         `json:"name,omitempty"`
	Jobs []*WorkflowJob `json:"jobs,omitempty"`
}

type WorkflowJob struct {
	Name  string      `json:"name,omitempty"`
	Steps []*TaskStep `json:"steps,omitempty"`
}

// TaskStep represents a single task step from a sequence of tasks of a job.
type TaskStep struct {
	Name             string `json:"name,omitempty"`
	Run              string `json:"run,omitempty"`
	Uses             string `json:"run,omitempty"`
	WorkingDirectory string `json:"working-directory,omitempty"`
}
