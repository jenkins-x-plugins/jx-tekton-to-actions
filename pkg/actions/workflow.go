package actions

// Workflow represents a github workflow
type Workflow struct {
	Name     string                  `json:"name,omitempty"`
	On       Events                  `json:"on,omitempty"`
	Env      map[string]string       `json:"env,omitempty"`
	Defaults *Defaults               `json:"defaults,omitempty"`
	Jobs     map[string]*WorkflowJob `json:"jobs,omitempty"`
}

type WorkflowJob struct {
	RunsOn string      `json:"runs-on,omitempty"`
	Steps  []*TaskStep `json:"steps,omitempty"`
}

// TaskStep represents a single task step from a sequence of tasks of a job.
type TaskStep struct {
	Name             string            `json:"name,omitempty"`
	Run              string            `json:"run,omitempty"`
	Shell            string            `json:"shell,omitempty"`
	Uses             string            `json:"uses,omitempty"`
	With             map[string]string `json:"with,omitempty"`
	WorkingDirectory string            `json:"working-directory,omitempty"`
}

type Defaults struct {
	Shell            string `json:"shell,omitempty"`
	WorkingDirectory string `json:"working-directory,omitempty"`
}
