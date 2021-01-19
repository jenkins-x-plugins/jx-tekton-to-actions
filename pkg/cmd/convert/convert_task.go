package convert

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/actions"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

const (
	stripUseBinEnv              = true
	replaceWorkspaceSourcePaths = false
)

var (
	stepEnvironmentVariables = map[string]map[string]string{
		/*
			"promote-helm-release": {
				"JX_REPOSITORY_USERNAME": "${{ github.repository_owner }}",
				"JX_REPOSITORY_PASSWORD": "${{ secrets.GITHUB_TOKEN }}",
			},
		*/
	}
)

func (o *Options) taskToJob(spec *v1beta1.TaskSpec, kind TriggerKind, dirName string) (*actions.WorkflowJob, error) {
	checkout := &actions.TaskStep{
		Name: "Checkout",
		Uses: "actions/checkout@v2",
	}
	if kind == TriggerPostsubmit {
		// lets do a full clone for tags
		checkout.With = map[string]string{
			"fetch-depth": "0",
		}
	}
	job := &actions.WorkflowJob{
		RunsOn: o.RunsOn,
		Steps: []*actions.TaskStep{
			checkout,
		},
	}
	for i := range spec.Steps {
		s := &spec.Steps[i]
		if stringhelpers.StringArrayIndex(o.RemoveSteps, s.Name) >= 0 {
			continue
		}
		taskSteps, err := o.taskStepToTaskStep(spec, s, kind, dirName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create step for %s", s.Name)
		}

		job.Steps = append(job.Steps, taskSteps...)
	}
	return job, nil
}

func (o *Options) taskStepToTaskStep(spec *v1beta1.TaskSpec, s *v1beta1.Step, kind TriggerKind, dirName string) ([]*actions.TaskStep, error) {
	override := o.Overrides[s.Name]
	if override != nil {
		args := &StepOverrideArgs{
			Name: s.Name,
			Kind: kind,
		}
		return override(args), nil
	}
	step := &actions.TaskStep{
		Name: s.Name,
		Uses: "docker://" + s.Image,
		With: map[string]string{},
		Env: map[string]string{
			"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
		},
	}
	for k, v := range stepEnvironmentVariables[s.Name] {
		step.Env[k] = v
	}
	script := s.Script
	if script != "" {
		if replaceWorkspaceSourcePaths {
			// lets remove all absolute paths to .jx
			script = strings.ReplaceAll(script, "/workspace/source/", "")
		}

		// lets get the first line
		i := strings.Index(script, "\n")
		if i > 0 {
			shebangLine := strings.TrimSpace(script[0:i])
			if strings.HasPrefix(shebangLine, shebang) {
				shell := strings.TrimPrefix(shebangLine, shebang)

				remaining := strings.TrimSpace(script[i+1:])
				lines := strings.Split(remaining, "\n")
				if len(lines) == 1 && strings.HasPrefix(lines[0], "jx ") {
					line := strings.TrimPrefix(lines[0], "jx ")
					replacement := replacements[line]
					if replacement != "" {
						line = replacement
					}
					step.With["args"] = line
					step.With["entrypoint"] = "jx"
				} else if len(lines) > 2 {

					// lets create a script and use that
					fileName := s.Name + ".sh"
					dir := filepath.Join(o.OutDir, dirName)
					path := filepath.Join(dir, fileName)
					err := os.MkdirAll(dir, files.DefaultDirWritePermissions)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to create dir %s", dir)
					}
					err = ioutil.WriteFile(path, []byte(script), 0777)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to save file %s", path)
					}
					step.With["entrypoint"] = path
					log.Logger().Infof("created file %s", info(path))
				} else {
					remaining = strings.ReplaceAll(remaining, "\n", "; ")
					remaining = strings.ReplaceAll(remaining, `"`, `\"`)
					remaining = strings.ReplaceAll(remaining, `'`, `\'`)

					replacement := replacements[remaining]
					if replacement != "" {
						remaining = replacement
					}

					args := "-c \"" + remaining + "\""
					if strings.HasPrefix(shell, "/usr/bin/env ") {
						args = strings.TrimPrefix(shell, "/usr/bin/env ") + " " + args
						shell = "/usr/bin/env"
						if stripUseBinEnv {
							s := strings.SplitN(args, " ", 2)
							if len(s) == 2 {
								shell = s[0]
								args = s[1]
							}
						}
					}
					step.With["entrypoint"] = shell
					step.With["args"] = args
				}
			}
		}
	} else {
		step.With["entrypoint"] = strings.Join(append(s.Command, s.Args...), " ")
	}
	return []*actions.TaskStep{step}, nil
}
