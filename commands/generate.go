package commands

import (
	"errors"
	"path"

	"github.com/contextcloud/ccb-cli/pkg/deployer"
	"github.com/contextcloud/ccb-cli/pkg/parser"
	"github.com/contextcloud/ccb-cli/pkg/print"

	"github.com/spf13/cobra"
)

var (
	// ErrNoMetadata when a file does not contain metadata
	ErrNoMetadata = errors.New("No metadata found in file")
	// ErrNoSpec when a file does not contain spec
	ErrNoSpec = errors.New("No spec found in file")
	// ErrInvalidKind kind is not support
	ErrInvalidKind = errors.New("Unsupported kind")
	// ErrInvalidNamespace when two namespaces don't match
	ErrInvalidNamespace = errors.New("Namespaces don't match")
	// ErrNoConfig when the config isn't supplied
	ErrNoConfig = errors.New("No config supplied")
)

type generateOptions struct {
	stackFile  string
	workingDir string
	tag        string
	registry   string
	namespace  string
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
		ccb generate -f https://domain/path/service.yml
		ccb generate -f ./service.yml`,
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

	de := deployer.NewManager(opts.workingDir, opts.namespace, opts.registry, opts.tag)
	for _, fn := range fns {
		f := &deployer.Function{
			Key:      fn.Key,
			Name:     fn.Name,
			Version:  fn.Version,
			Env:      fn.Env,
			Envs:     fn.Envs,
			Secrets:  fn.Secrets,
			Limits:   fn.Limits,
			Requests: fn.Requests,
		}

		de.AddFunction(f)
	}

	manifests, err := de.Generate()
	if err != nil {
		return err
	}

	if opts.output != "" {
		return manifests.Save(opts.output)
	}

	manifests.Print(env.Out)
	return nil
}
