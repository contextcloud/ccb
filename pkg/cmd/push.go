package cmd

import (
	"context"
	"fmt"

	"github.com/contextcloud/ccb-cli/pkg/builder"
	"github.com/contextcloud/ccb-cli/pkg/spec"

	"github.com/spf13/cobra"
)

// pushCmd represents the pack command
var pushCmd = &cobra.Command{
	Use:   `push`,
	Short: "Push will build and push docker images",
	Long:  `Push will build all functions after they have been packed and push the image to a registry`,
	Example: `
  ccb push -f https://domain/path/service.yml
  ccb push -f ./service.yml`,
	RunE: runPush,
}

func init() {
	pushCmd.Flags().StringVarP(&tag, "tag", "t", defaultTag, "Override latest tag on function Docker image")
	pushCmd.Flags().StringVar(&registry, "registry", defaultRegistry, "The registry to find the Docker Image")

	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
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

	pushed, err := b.Push(context.Background())
	if err != nil {
		return err
	}
	for _, r := range pushed {
		fmt.Printf("Push %s\n", r)
	}
	return nil
}
