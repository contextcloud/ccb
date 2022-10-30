package commands

import (
	"path"

	"github.com/contextcloud/ccb/pkg/deployer"
	"github.com/contextcloud/ccb/pkg/parser"
	"github.com/contextcloud/ccb/pkg/print"

	"github.com/spf13/cobra"
)

type routesOptions struct {
	stackFile  string
	workingDir string
	namespace  string
	commit     string
	output     string
}

func newRoutesCommand() *cobra.Command {
	logger := print.NewConsoleLogger()
	options := routesOptions{}

	// generateCmd represents the generate command
	cmd := &cobra.Command{
		Use:   `routes`,
		Short: "generates http proxy routes for Kubernetes",
		Long:  `generates http proxy routes Manifest files using a spec provided in yaml`,
		Example: `
		ccb routes -f https://domain/path/stack.yml
		ccb routes -f ./stack.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRoutes(logger, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")
	flags.StringVarP(&options.namespace, "namespace", "n", "", "The network to connect to")
	flags.StringVarP(&options.commit, "commit", "", "", "The commit label")
	flags.StringVarP(&options.output, "output", "o", "", "Where to save the files")

	return cmd
}

func runRoutes(logger print.Logger, opts routesOptions, args []string) error {
	stackFile := path.Join(opts.workingDir, opts.stackFile)

	stack, err := parser.LoadStack(stackFile)
	if err != nil {
		return err
	}

	routes, err := stack.GetRoutes(args...)
	if err != nil {
		return err
	}

	if len(routes) == 0 {
		logger.Err().Println("No routes found")
		return nil
	}

	de := deployer.NewManager(opts.workingDir, opts.namespace, opts.commit)

	manifests, err := de.GenerateRoutes(routes)
	if err != nil {
		return err
	}

	if opts.output != "" {
		return manifests.Save(opts.output)
	}

	manifests.Print(logger.Out())
	return nil
}
