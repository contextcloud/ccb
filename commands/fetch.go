package commands

import (
	"context"
	"path"

	"github.com/contextcloud/ccb/pkg/parser"
	"github.com/contextcloud/ccb/pkg/print"
	"github.com/contextcloud/ccb/pkg/templater"

	"github.com/spf13/cobra"
)

type fetchOptions struct {
	stackFile  string
	workingDir string
}

func newFetchCommand() *cobra.Command {
	env := print.NewEnv()
	options := fetchOptions{}

	// fetchCmd represents the pack command
	cmd := &cobra.Command{
		Use:   `fetch`,
		Short: "fetch downloads all templates",
		Long:  `fetch finds all templates and downloads them`,
		Example: `
  ccb fetch -f https://domain/path/stack.yml
  ccb fetch -f ./stack.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetch(env, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")

	return cmd
}

func runFetch(env *print.Env, opts fetchOptions, args []string) error {
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
		env.Err.Println("Download failed: ", err)
		return err
	}

	for _, path := range downloaded {
		env.Out.Println("Fetched", path)
	}
	return nil
}
