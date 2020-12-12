package cmd

import (
	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/cmd/convert"
	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/cmd/version"
	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/rootcmd"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
)

// Main creates the new command
func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootcmd.TopLevelCommand,
		Short: "commands for converting tekton pipelines to github actions",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Logger().Errorf(err.Error())
			}
		},
	}

	cmd.AddCommand(cobras.SplitCommand(convert.NewCmdConvert()))
	cmd.AddCommand(cobras.SplitCommand(version.NewCmdVersion()))
	return cmd
}
