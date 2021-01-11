package convert

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/actions"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-helpers/v3/pkg/yamls"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/jenkins-x/lighthouse/pkg/config/job"
	"github.com/jenkins-x/lighthouse/pkg/triggerconfig"
	"github.com/jenkins-x/lighthouse/pkg/triggerconfig/inrepo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type TriggerKind string

const (
	// TriggerPresubmit for presubmits
	TriggerPresubmit TriggerKind = "presubmit"

	// TriggerPostsubmit for postsubmits
	TriggerPostsubmit TriggerKind = "postsubmit"
)

type StepOverrideFunction func(*StepOverrideArgs) []*actions.TaskStep

type StepOverrideArgs struct {
	Name string
	Kind TriggerKind
}

// Options contains the command line options
type Options struct {
	options.BaseOptions

	Dir          string
	OutDir       string
	RunsOn       string
	MainBranches []string
	RemoveSteps  []string
	Recursive    bool

	Workflows        map[string]*actions.Workflow
	Overrides        map[string]StepOverrideFunction
	LoginToDockerHub bool
}

var (
	defaultMainBranches = []string{"main", "master"}
	defaultRemoveSteps  = []string{"git-clone", "setup-builder-home", "git-merge"}

	shebang = "#!"

	info = termcolor.ColorInfo

	cmdLong = templates.LongDesc(`
		Converts tekton pipelines to github actions
`)

	cmdExample = templates.Examples(`
		# Converts the tekton pipelines to actions
		jx tekton-to-actions convert
	`)

	replacements = map[string]string{
		// lets workaround git commit not yet working inside GHA until we lazy add git config
		"jx gitops variables": "jx gitops variables --commit=false",

		// disable the kaniko copy for now
		"source .jx/variables.sh; cp /tekton/creds-secrets/tekton-container-registry-auth/.dockerconfigjson /kaniko/.docker/config.json; /kaniko/executor $KANIKO_FLAGS --context=/workspace/source --dockerfile=/workspace/source/Dockerfile --destination=$DOCKER_REGISTRY/$DOCKER_REGISTRY_ORG/$APP_NAME:$VERSION": "source .jx/variables.sh; /kaniko/executor $KANIKO_FLAGS --context=. --dockerfile=Dockerfile --destination=$DOCKER_REGISTRY/$DOCKER_REGISTRY_ORG/$APP_NAME:$VERSION",
	}
)

// NewCmdConvert creates the command
func NewCmdConvert() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "convert",
		Short:   "Converts tekton pipelines to github actions",
		Long:    cmdLong,
		Example: cmdExample,
		Aliases: []string{"kill"},
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&o.Dir, "dir", "d", ".", "The directory to look for the .lighthouse folder")
	cmd.Flags().StringVarP(&o.OutDir, "output-dir", "o", "", "The directory to write output files")
	cmd.Flags().StringVarP(&o.RunsOn, "runs-on", "", "ubuntu-latest", "The machine this runs on")
	cmd.Flags().StringArrayVarP(&o.MainBranches, "main-branches", "", defaultMainBranches, "The main branches for releases")
	cmd.Flags().StringArrayVarP(&o.RemoveSteps, "remove-steps", "", defaultRemoveSteps, "The steps to remove")
	cmd.Flags().BoolVarP(&o.Recursive, "recursive", "r", false, "Recursively find all '.lighthouse' folders such as if linting a Pipeline Catalog")
	return cmd, o
}

// Run implements this command
func (o *Options) Run() error {
	if o.OutDir == "" {
		o.OutDir = filepath.Join(o.Dir, ".github", "workflows")
	}
	err := os.MkdirAll(o.OutDir, files.DefaultDirWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to create output dir %s", o.OutDir)
	}

	if o.Overrides == nil {
		o.Overrides = o.defaultOverrides()
	}

	if o.Recursive {
		err := filepath.Walk(o.Dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info == nil || !info.IsDir() || info.Name() != ".lighthouse" {
				return nil
			}
			return o.ProcessDir(path)
		})
		if err != nil {
			return err
		}
	} else {
		dir := filepath.Join(o.Dir, ".lighthouse")
		err := o.ProcessDir(dir)
		if err != nil {
			return err
		}
	}

	return o.writeWorkflows()
}

func (o *Options) ProcessDir(dir string) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to read dir %s", dir)
	}
	for _, f := range fs {
		name := f.Name()
		if !f.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		triggerDir := filepath.Join(dir, name)
		triggersFile := filepath.Join(triggerDir, "triggers.yaml")
		exists, err := files.FileExists(triggersFile)
		if err != nil {
			return errors.Wrapf(err, "failed to check if file exists %s", triggersFile)
		}
		if !exists {
			continue
		}

		triggers := &triggerconfig.Config{}
		err = yamls.LoadFile(triggersFile, triggers)
		if err != nil {
			return errors.Wrapf(err, "failed to load triggers file %s", triggersFile)
		}
		err = o.processTriggers(triggers, triggerDir, name)
		if err != nil {
			return errors.Wrapf(err, "failed to process triggers file %s", triggersFile)
		}
	}
	return nil
}

func (o *Options) processTriggers(repoConfig *triggerconfig.Config, dir string, name string) error {
	ctx := context.TODO()
	for i := range repoConfig.Spec.Presubmits {
		r := &repoConfig.Spec.Presubmits[i]
		if r.SourcePath != "" {
			path := filepath.Join(dir, r.SourcePath)
			pr, err := loadJobBaseFromSourcePath(ctx, path)
			if err != nil {
				return errors.Wrapf(err, "failed to load pipeline at %s", path)
			}
			r.PipelineRunSpec = &pr.Spec
			events := o.presubmitToEvents(r)
			err = o.processTriggerPipeline(repoConfig, &r.Base, name, events, TriggerPresubmit)
			if err != nil {
				return errors.Wrapf(err, "failed to process pipeline at %s", path)
			}
		}
	}
	for i := range repoConfig.Spec.Postsubmits {
		r := &repoConfig.Spec.Postsubmits[i]
		if r.SourcePath != "" {
			path := filepath.Join(dir, r.SourcePath)
			pr, err := loadJobBaseFromSourcePath(ctx, path)
			if err != nil {
				return errors.Wrapf(err, "failed to load pipeline at %s", path)
			}
			r.PipelineRunSpec = &pr.Spec
			events := o.postsubmitToEvents(r)
			err = o.processTriggerPipeline(repoConfig, &r.Base, name, events, TriggerPostsubmit)
			if err != nil {
				return errors.Wrapf(err, "failed to process pipeline at %s", path)
			}
		}
	}
	return nil
}

func (o *Options) presubmitToEvents(r *job.Presubmit) actions.Events {
	answer := actions.Events{
		PullRequest: &actions.BranchEvent{},
		Push: &actions.BranchEvent{
			BranchesIgnore: o.MainBranches,
		},
	}
	return answer
}

func (o *Options) postsubmitToEvents(r *job.Postsubmit) actions.Events {
	answer := actions.Events{
		Push: &actions.BranchEvent{
			Branches: o.MainBranches,
		},
	}
	return answer
}

func (o *Options) processTriggerPipeline(config *triggerconfig.Config, jobBase *job.Base, name string, events actions.Events, kind TriggerKind) error {
	prSpec := jobBase.PipelineRunSpec
	if prSpec == nil || prSpec.PipelineSpec == nil {
		return nil
	}

	fileName := name + "-" + jobBase.Name + ".yaml"
	workflow := &actions.Workflow{
		On: events,
	}
	for _, pt := range prSpec.PipelineSpec.Tasks {
		if pt.TaskSpec == nil || pt.TaskSpec.TaskSpec == nil {
			continue
		}
		job := o.taskToJob(pt.TaskSpec.TaskSpec, kind)
		if job != nil {
			if workflow.Jobs == nil {
				workflow.Jobs = map[string]*actions.WorkflowJob{}
			}
			workflow.Jobs[jobBase.Name] = job
		}
	}
	if o.Workflows == nil {
		o.Workflows = map[string]*actions.Workflow{}
	}
	o.Workflows[fileName] = workflow
	return nil
}

func (o *Options) taskToJob(spec *v1beta1.TaskSpec, kind TriggerKind) *actions.WorkflowJob {
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
		taskSteps := o.taskStepToTaskStep(spec, s, kind)
		job.Steps = append(job.Steps, taskSteps...)
	}
	return job
}

func (o *Options) taskStepToTaskStep(spec *v1beta1.TaskSpec, s *v1beta1.Step, kind TriggerKind) []*actions.TaskStep {
	override := o.Overrides[s.Name]
	if override != nil {
		args := &StepOverrideArgs{
			Name: s.Name,
			Kind: kind,
		}
		return override(args)
	}
	step := &actions.TaskStep{
		Name: s.Name,
		Uses: "docker://" + s.Image,
		With: map[string]string{},
		Env: map[string]string{
			"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
		},
	}
	if s.Script != "" {
		// lets get the first line
		i := strings.Index(s.Script, "\n")
		if i > 0 {
			shebangLine := strings.TrimSpace(s.Script[0:i])
			if strings.HasPrefix(shebangLine, shebang) {
				shell := strings.TrimPrefix(shebangLine, shebang)

				remaining := strings.TrimSpace(s.Script[i+1:])
				lines := strings.Split(remaining, "\n")
				if len(lines) == 1 && strings.HasPrefix(lines[0], "jx ") {
					line := lines[0]
					replacement := replacements[line]
					if replacement != "" {
						line = replacement
					}
					step.With["args"] = line
				} else {
					remaining = strings.ReplaceAll(remaining, "\n", "; ")
					remaining = strings.ReplaceAll(remaining, `"`, `\"`)
					remaining = strings.ReplaceAll(remaining, `'`, `\'`)

					replacement := replacements[remaining]
					if replacement != "" {
						remaining = replacement
					}

					if shell == "/busybox/sh" {
						step.With["entrypoint"] = shell
						step.With["args"] = "-c \"" + remaining + "\""
					} else {
						step.With["args"] = shell + " -c \"" + remaining + "\""
					}
				}
			}
		}
	} else {
		step.With["entrypoint"] = strings.Join(append(s.Command, s.Args...), " ")
	}
	return []*actions.TaskStep{step}
}

func loadJobBaseFromSourcePath(ctx context.Context, path string) (*v1beta1.PipelineRun, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load file %s", path)
	}
	if len(data) == 0 {
		return nil, errors.Errorf("empty file file %s", path)
	}

	dir := filepath.Dir(path)
	message := fmt.Sprintf("file %s", path)

	getData := func(path string) ([]byte, error) {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read file %s", path)
		}
		return data, nil
	}

	pr, err := inrepo.LoadTektonResourceAsPipelineRun(data, dir, message, getData, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal YAML file %s", path)
	}
	return pr, nil
}

func (o *Options) writeWorkflows() error {
	for f, w := range o.Workflows {
		path := filepath.Join(o.OutDir, f)
		err := yamls.SaveFile(w, path)
		if err != nil {
			return errors.Wrapf(err, "failed to save file %s", path)
		}
		log.Logger().Infof("saved file %s", info(path))
	}
	return nil
}

func (o *Options) defaultOverrides() map[string]StepOverrideFunction {
	return map[string]StepOverrideFunction{
		"build-container-build": o.createDockerBuildStep,
	}
}

func (o *Options) createDockerBuildStep(args *StepOverrideArgs) []*actions.TaskStep {
	answer := []*actions.TaskStep{
		{
			Name: "Set up QEMU",
			Uses: "docker/setup-qemu-action@v1",
		},
		{
			Name: "Set up Docker Buildx",
			Uses: "docker/setup-buildx-action@v1",
		},
	}
	if o.LoginToDockerHub {
		answer = append(answer, &actions.TaskStep{
			Name: "Login to DockerHub",
			Uses: "docker/login-action@v1",
			With: map[string]string{
				"username": `${{ secrets.DOCKERHUB_USERNAME }}`,
				"password": `${{ secrets.${{ secrets.DOCKERHUB_USERNAME }} }}`,
			},
		})
	}
	pushWith := map[string]string{
		"context": ".",
		"file":    "./Dockerfile",
	}

	/* if release....
		      platforms: linux/amd64,linux/arm64,linux/386
	          push: true
	          tags: user/app:latest
	*/

	answer = append(answer, &actions.TaskStep{
		Name: "Build and push",
		Uses: "docker/build-push-action@v2",
		With: pushWith,
	})
	return answer

}
