package cmd

import (
	"context"
	"fmt"

	"github.com/contextgg/faas-cd/templater"

	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
)

// fetchCmd represents the pack command
var fetchCmd = &cobra.Command{
	Use:   `fetch`,
	Short: "fetch downloads all templates",
	Long:  `fetch finds all templates and downloads them`,
	Example: `
  faas-cd fetch -f https://domain/path/service.yml
  faas-cd fetch -f ./service.yml`,
	RunE: runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}

func runFetch(cmd *cobra.Command, args []string) error {
	parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
	if err != nil {
		return err
	}

	var opts []templater.Option
	for _, templateSource := range parsedServices.StackConfiguration.TemplateConfigs {
		opts = append(opts, templater.AddLocationOption(templateSource.Name, templateSource.Source))
	}

	t := templater.NewTemplater(opts...)
	for name, fn := range parsedServices.Functions {
		// Need to fetch templates.
		t.AddFunction(name, fn.Language)
	}

	downloaded, err := t.Download(context.Background())
	if err != nil {
		return err
	}

	for _, path := range downloaded {
		fmt.Printf("Fetched %s", path)
	}
	return nil
}
