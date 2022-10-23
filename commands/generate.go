package commands

import (
	"path"

	"github.com/contextcloud/ccb/pkg/deployer"
	"github.com/contextcloud/ccb/pkg/parser"
	"github.com/contextcloud/ccb/pkg/print"

	"github.com/spf13/cobra"
)

type generateOptions struct {
	stackFile  string
	workingDir string
	tag        string
	registry   string
	namespace  string
	commit     string
	output     string
}

func newGenerateCommand() *cobra.Command {
	env := print.NewEnv()
	options := generateOptions{}

	// generateCmd represents the generate command
	cmd := &cobra.Command{
		Use:   `generate`,
		Short: "generates Kubernetes Manifests",
		Long:  `generates Kubernetes Manifest files using a spec provided in yaml`,
		Example: `
		ccb generate -f https://domain/path/stack.yml
		ccb generate -f ./stack.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(env, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")
	flags.StringVarP(&options.tag, "tag", "t", "latest", "The tag for the containers")
	flags.StringVarP(&options.registry, "registry", "", "", "The registry for the docker containers")
	flags.StringVarP(&options.namespace, "namespace", "n", "", "The network to connect to")
	flags.StringVarP(&options.commit, "commit", "", "", "The commit label")
	flags.StringVarP(&options.output, "output", "o", "", "Where to save the files")

	return cmd
}

func runGenerate(env *print.Env, opts generateOptions, args []string) error {
	stackFile := path.Join(opts.workingDir, opts.stackFile)

	stack, err := parser.LoadStack(stackFile)
	if err != nil {
		return err
	}

	fns, err := stack.GetFunctions(args...)
	if err != nil {
		return err
	}

	if len(fns) == 0 {
		env.Err.Println("No functions found")
		return nil
	}

	de := deployer.NewManager(opts.workingDir, opts.namespace, opts.commit)

	manifests, err := de.GenerateFunctions(opts.registry, opts.tag, fns)
	if err != nil {
		return err
	}

	if opts.output != "" {
		return manifests.Save(opts.output)
	}

	manifests.Print(env.Out)
	return nil
}
