package commands

import (
	"os"

	"github.com/spf13/cobra"
)

const defaultStackFile = "stack.yml"
const defaultWorkingDir = "."
const rootCommandName = "ccb"

// These variables are initialized externally during the build. See the Makefile.
var Version string
var GitCommit string

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootCommandName,
		Short: "Manage your Context Cloud functions",
		Long:  `	Manage your Context Cloud functions from the command line`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()
			root.Version = Version
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
		},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newBuildCommand())
	cmd.AddCommand(newFetchCommand())
	cmd.AddCommand(newGenerateCommand())
	cmd.AddCommand(newRoutesCommand())
	cmd.AddCommand(newVersionCommand())

	return cmd
}
