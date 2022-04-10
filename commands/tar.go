package commands

import (
	"context"
	"path"

	"github.com/contextcloud/ccb-cli/pkg/parser"
	"github.com/contextcloud/ccb-cli/pkg/print"
	"github.com/contextcloud/ccb-cli/pkg/templater"

	"github.com/spf13/cobra"
)

type tarOptions struct {
	stackFile  string
	workingDir string
}

func newTarCommand() *cobra.Command {
	env := print.NewEnv()
	options := tarOptions{}

	// packCmd represents the pack command
	cmd := &cobra.Command{
		Use:   `tar`,
		Short: "tar tarball for build context",
		Long:  `tar converts functions into tarballs`,
		Example: `
		ccb tar -f https://domain/path/stack.yml
		ccb tar -f ./stack.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTar(env, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")

	return cmd
}

func runTar(env *print.Env, opts tarOptions, args []string) error {
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

	t := templater.NewTemplater(opts.workingDir)
	for _, fn := range fns {
		// Need to fetch templates.
		t.AddFunction(fn.Key, fn.Template)
	}

	downloaded, err := t.Download(context.Background())
	if err != nil {
		return err
	}
	for _, path := range downloaded {
		env.Out.Println("Fetched:", path)
	}

	tars, err := t.Tar(context.Background())
	if err != nil {
		return err
	}
	for _, path := range tars {
		env.Out.Println("Tar:", path)
	}

	return nil
}
