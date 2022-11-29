package commands

import (
	"context"
	"path"

	"github.com/contextcloud/ccb/pkg/parser"
	"github.com/contextcloud/ccb/pkg/print"
	"github.com/contextcloud/ccb/pkg/templater"
	"github.com/contextcloud/ccb/pkg/utils"

	"github.com/spf13/cobra"
)

type fetchOptions struct {
	stackFile  string
	workingDir string
}

func newFetchCommand() *cobra.Command {
	logger := print.NewConsoleLogger()
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
			return runFetch(logger, options, args)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&options.stackFile, "stack", "f", defaultStackFile, "Path to Stack file")
	flags.StringVarP(&options.workingDir, "working-dir", "d", defaultWorkingDir, "Working directory")

	return cmd
}

func runFetch(logger print.Logger, opts fetchOptions, args []string) error {
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

	t := templater.NewTemplater(opts.workingDir)
	for _, fn := range fns {
		if utils.IsDockerTemplate(fn.Template) {
			continue
		}

		// Need to fetch templates.
		t.AddFunction(fn.Key, fn.Template)
	}

	ctx := context.Background()
	downloaded, err := t.Download(ctx)
	if err != nil {
		logger.Err().Println("Download failed: ", err)
		return err
	}

	for _, path := range downloaded {
		logger.Out().Println("Fetched", path)
	}
	return nil
}
