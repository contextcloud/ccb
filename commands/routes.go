package commands

import (
	"path"

	"github.com/contextcloud/ccb-cli/pkg/deployer"
	"github.com/contextcloud/ccb-cli/pkg/parser"
	"github.com/contextcloud/ccb-cli/pkg/print"

	"github.com/spf13/cobra"
)

type routesOptions struct {
	stackFile  string
	workingDir string
	namespace  string
	output     string
}

func newRoutesCommand() *cobra.Command {
	env := print.NewEnv()
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
			return runRoutes(env, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")
	flags.StringVarP(&options.namespace, "namespace", "n", "", "The network to connect to")
	flags.StringVarP(&options.output, "output", "o", "", "Where to save the files")

	return cmd
}

func runRoutes(env *print.Env, opts routesOptions, args []string) error {
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
		env.Err.Println("No routes found")
		return nil
	}

	all := make([]*deployer.Route, len(routes))
	for i, r := range routes {
		all[i] = &deployer.Route{
			Key:      r.Key,
			FQDN:     r.FQDN,
			Includes: r.Includes,
		}
	}

	de := deployer.NewManager(opts.workingDir, opts.namespace)

	manifests, err := de.GenerateRoutes(all)
	if err != nil {
		return err
	}

	if opts.output != "" {
		return manifests.Save(opts.output)
	}

	manifests.Print(env.Out)
	return nil
}
