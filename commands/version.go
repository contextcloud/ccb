package commands

import (
	"runtime"

	"github.com/spf13/cobra"

	"github.com/contextcloud/ccb/pkg/print"
)

type versionOptions struct {
	all bool
}

func newVersionCommand() *cobra.Command {
	env := print.NewEnv()
	options := versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show ccb version information.",
		Run: func(cmd *cobra.Command, args []string) {
			runVersion(env, options, cmd.Root())
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.BoolVarP(&options.all, "all", "a", false,
		"Show all version information",
	)

	return cmd
}

func runVersion(env *print.Env, opts versionOptions, root *cobra.Command) {
	if opts.all {
		env.Out.Printf("%s version: %s\n", rootCommandName, root.Version)
		env.Out.Printf("System version: %s/%s\n", runtime.GOARCH, runtime.GOOS)
		env.Out.Printf("Golang version: %s\n", runtime.Version())
		return
	}

	env.Out.Printf("%s version: %s\n", rootCommandName, root.Version)
}
