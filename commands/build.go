package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"path"

	"github.com/contextcloud/ccb/pkg/builder"
	"github.com/contextcloud/ccb/pkg/parser"
	"github.com/contextcloud/ccb/pkg/print"
	"github.com/contextcloud/ccb/pkg/utils"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/docker/api/types"

	"github.com/spf13/cobra"
)

type buildOptions struct {
	stackFile  string
	workingDir string
	network    string
	buildArgs  []string

	push bool

	tag      string
	registry string
	prefix   string
	username string
	password string

	poolSize int
}

func newBuildCommand() *cobra.Command {
	logger := print.NewConsoleLogger()
	options := buildOptions{}

	// buildCmd represents the pack command
	cmd := &cobra.Command{
		Use:   `build`,
		Short: "build makes docker images",
		Long:  `build makes docker images from our packed dirs`,
		Example: `
  ccb build -f https://domain/path/stack.yml
  ccb build -f ./stack.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(logger, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")
	flags.StringVarP(&options.network, "network", "", "", "The network to connect to")
	flags.StringSliceVarP(&options.buildArgs, "build-args", "", []string{}, "To be parsed as a key=value pair to docker build")

	flags.BoolVarP(&options.push, "push", "", false, "If true, push images to registry")

	flags.StringVarP(&options.tag, "tag", "t", "latest", "The tag for the containers")
	flags.StringVarP(&options.registry, "registry", "", "", "The registry for the docker images")
	flags.StringVarP(&options.prefix, "prefix", "", "", "The prefix for the docker image name")
	flags.StringVarP(&options.username, "username", "", "", "The username for the registry")
	flags.StringVarP(&options.password, "password", "", "", "The password for the registry")

	flags.IntVarP(&options.poolSize, "pool-size", "", 1, "How many containers to build together")

	return cmd
}

func runBuild(logger print.Logger, opts buildOptions, args []string) error {
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
		logger.Err().Println("No functions found")
		return nil
	}

	gargs, err := utils.ParseMap(opts.buildArgs, "build-args")
	if err != nil {
		return err
	}

	registry := "https://index.docker.io/v1/"
	if opts.registry != "" {
		registry = opts.registry
	}

	cfg, err := cliconfig.Load("")
	if err != nil {
		panic(err)
	}

	a, _ := cfg.GetAuthConfig(registry)
	ac := types.AuthConfig(a)
	if opts.username != "" {
		ac.Username = opts.username
	}
	if opts.password != "" {
		ac.Password = opts.password
	}

	authConfigBytes, _ := json.Marshal(ac)
	registryAuth := base64.URLEncoding.EncodeToString(authConfigBytes)

	buildOptions := &builder.Options{
		Log:        logger.Out(),
		WorkingDir: opts.workingDir,
		PoolSize:   opts.poolSize,
		Network:    opts.network,

		Push:         opts.push,
		Registry:     opts.registry,
		Prefix:       opts.prefix,
		Tag:          opts.tag,
		RegistryAuth: registryAuth,
	}
	b, err := builder.NewBuilder(buildOptions)
	if err != nil {
		return err
	}

	for _, fn := range fns {
		// Args!
		args := utils.MergeMap(gargs, fn.BuildArgs)

		// Need to fetch templates.
		b.AddService(fn.Key, fn.Template, args)
	}

	built, err := b.Build(context.Background())
	if err != nil {
		return err
	}

	for _, r := range built {
		logger.Out().Println("Built", r)
	}
	return nil
}
