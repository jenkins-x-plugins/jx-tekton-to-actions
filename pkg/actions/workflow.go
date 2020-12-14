package actions

// Workflow represents a github workflow
type Workflow struct {
	Name     string            `json:"name,omitempty"`
	On       Events            `json:"on,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Defaults *Defaults         `json:"defaults,omitempty"`
	Jobs     []*WorkflowJob    `json:"jobs,omitempty"`
}

type WorkflowJob struct {
	Name  string      `json:"name,omitempty"`
	Steps []*TaskStep `json:"steps,omitempty"`
}

// TaskStep represents a single task step from a sequence of tasks of a job.
type TaskStep struct {
	Name             string `json:"name,omitempty"`
	Run              string `json:"run,omitempty"`
	Uses             string `json:"uses,omitempty"`
	WorkingDirectory string `json:"working-directory,omitempty"`
}

type Defaults struct {
	Shell            string `json:"shell,omitempty"`
	WorkingDirectory string `json:"working-directory,omitempty"`
}
