package cmd

import (
	"context"
	"fmt"

	"github.com/contextcloud/ccb-cli/builder"
	"github.com/contextcloud/ccb-cli/spec"

	"github.com/spf13/cobra"
)

// buildCmd represents the pack command
var buildCmd = &cobra.Command{
	Use:   `build`,
	Short: "build makes docker images",
	Long:  `build makes docker images from our packed dirs`,
	Example: `
  ccb build -f https://domain/path/service.yml
  ccb build -f ./service.yml`,
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&tag, "tag", "t", defaultTag, "Override latest tag on function Docker image")
	buildCmd.Flags().StringVar(&registry, "registry", defaultRegistry, "The registry to find the Docker Image")

	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	parsedServices, err := spec.ParseYAMLFile(yamlFile, regex, filter, envsubst)
	if err != nil {
		return err
	}

	gargs, err := parseMap(buildArgs, "build-args")
	if err != nil {
		return err
	}

	b := builder.NewBuilder(
		builder.SetRegistry(registry),
		builder.SetTag(tag),
		builder.SetNetwork(network),
	)
	for name, val := range parsedServices.Functions {
		// Args!
		args := mergeMap(gargs, val.BuildArgs)

		// Need to fetch templates.
		b.AddService(name, args)
	}

	built, err := b.Build(context.Background())
	if err != nil {
		return err
	}
	for _, r := range built {
		fmt.Printf("Built %s\n", r)
	}

	return nil
}
