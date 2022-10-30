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
	logger := print.NewConsoleLogger()
	options := versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show ccb version information.",
		Run: func(cmd *cobra.Command, args []string) {
			runVersion(logger.Out(), options, cmd.Root())
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.BoolVarP(&options.all, "all", "a", false,
		"Show all version information",
	)

	return cmd
}

func runVersion(log print.Log, opts versionOptions, root *cobra.Command) {
	if opts.all {
		log.Printf("%s version: %s\n", rootCommandName, root.Version)
		log.Printf("System version: %s/%s\n", runtime.GOARCH, runtime.GOOS)
		log.Printf("Golang version: %s\n", runtime.Version())
		return
	}

	log.Printf("%s version: %s\n", rootCommandName, root.Version)
}
