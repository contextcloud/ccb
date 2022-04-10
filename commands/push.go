package commands

import (
	"context"
	"path"

	"github.com/contextcloud/ccb-cli/pkg/builder"
	"github.com/contextcloud/ccb-cli/pkg/parser"
	"github.com/contextcloud/ccb-cli/pkg/print"
	"github.com/contextcloud/ccb-cli/pkg/utils"

	"github.com/spf13/cobra"
)

func newPushCommand() *cobra.Command {
	env := print.NewEnv()
	options := buildOptions{}

	// pushCmd represents the pack command
	cmd := &cobra.Command{
		Use:   `push`,
		Short: "Push will build and push docker images",
		Long:  `Push will build all functions after they have been packed and push the image to a registry`,
		Example: `
  ccb push -f https://domain/path/stack.yml
  ccb push -f ./stack.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(env, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")
	flags.StringVarP(&options.tag, "tag", "t", "latest", "The tag for the containers")
	flags.StringVarP(&options.registry, "registry", "", "", "The registry for the docker containers")
	flags.StringVarP(&options.network, "network", "", "", "The network to connect to")
	flags.StringSliceVarP(&options.buildArgs, "build-args", "", []string{}, "To be parsed as a key=value pair to docker build")
	flags.IntVarP(&options.poolSize, "pool-size", "", 1, "How many containers to build together")

	return cmd
}

func runPush(env *print.Env, opts buildOptions, args []string) error {
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

	gargs, err := utils.ParseMap(opts.buildArgs, "build-args")
	if err != nil {
		return err
	}

	b := builder.NewBuilder(
		builder.SetRegistry(opts.registry),
		builder.SetTag(opts.tag),
		builder.SetNetwork(opts.network),
		builder.SetPoolSize(opts.poolSize),
	)

	for _, fn := range fns {
		// Args!
		args := utils.MergeMap(gargs, fn.BuildArgs)

		// Need to fetch templates.
		b.AddService(fn.Key, args)
	}

	built, err := b.Build(context.Background())
	if err != nil {
		return err
	}

	for _, r := range built {
		env.Out.Println("Built", r)
	}

	pushed, err := b.Push(context.Background())
	if err != nil {
		return err
	}
	for _, r := range pushed {
		env.Out.Println("Pushed", r)
	}
	return nil
}
