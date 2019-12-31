package cmd

import (
	"fmt"
	"strings"

	"github.com/contextgg/faas-cd/models"

	"github.com/openfaas/faas-cli/stack"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	defaultFunctionNamespace = "openfaas-fn"
	resourceKind             = "Function"
	defaultAPIVersion        = "openfaas.com/v1alpha2"
	defaultTag               = "latest"
	defaultRegistry          = ""
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   `generate`,
	Short: "generates Kubernetes Manifests",
	Long:  `generates Kubernetes Manifest files using a spec provided in yaml`,
	Example: `
  ccb generate -f https://domain/path/service.yml
  ccb generate -f ./service.yml`,
	RunE: runGenerate,
}

var (
	api               string
	functionNamespace string
	tag               string
	envsubst          bool
	registry          string
)

func init() {
	generateCmd.Flags().StringVar(&api, "api", defaultAPIVersion, "CRD API version e.g openfaas.com/v1alpha2, serving.knative.dev/v1alpha1")
	generateCmd.Flags().StringVarP(&functionNamespace, "namespace", "n", defaultFunctionNamespace, "Kubernetes namespace for functions")
	generateCmd.Flags().BoolVar(&envsubst, "envsubst", true, "Substitute environment variables in stack.yml file")
	generateCmd.Flags().StringVarP(&tag, "tag", "t", defaultTag, "Override latest tag on function Docker image")
	generateCmd.Flags().StringVar(&registry, "registry", defaultRegistry, "The registry to find the Docker Image")

	rootCmd.AddCommand(generateCmd)
}

func buildImageName(registry, imageName, tag string) string {
	tagger := func() string {
		ind := strings.Index(imageName, ":")
		if ind < 0 {
			return fmt.Sprintf("%s:%s", imageName, tag)
		}
		return fmt.Sprintf("%s:%s", imageName[:ind], tag)
	}

	if len(registry) > 0 {
		if strings.HasSuffix(registry, "/") {
			return fmt.Sprintf("%s/%s", registry[:len(registry)-1], tagger())
		}
		return fmt.Sprintf("%s/%s", registry, tagger())
	}
	return tagger()
}

func runGenerate(cmd *cobra.Command, args []string) error {
	parsedServices, err := stack.ParseYAMLFile(yamlFile, regex, filter, envsubst)
	if err != nil {
		return err
	}

	var manifests []models.Function
	for name, fn := range parsedServices.Functions {
		//read environment variables from the file
		fileEnvironment, err := readFiles(fn.EnvironmentFile)
		if err != nil {
			return err
		}

		//combine all environment variables
		allEnvironment, envErr := compileEnvironment([]string{}, fn.Environment, fileEnvironment)
		if envErr != nil {
			return envErr
		}

		var limits *models.FunctionResources
		if fn.Limits != nil {
			limits = &models.FunctionResources{
				Memory: fn.Limits.Memory,
				CPU:    fn.Limits.CPU,
			}
		}
		var requests *models.FunctionResources
		if fn.Requests != nil {
			requests = &models.FunctionResources{
				Memory: fn.Requests.Memory,
				CPU:    fn.Requests.CPU,
			}
		}
		var constraints []string
		if fn.Constraints != nil {
			constraints = *fn.Constraints
		}

		var environment *map[string]string
		if len(allEnvironment) > 0 {
			environment = &allEnvironment
		}

		spec := models.FunctionSpec{
			Name:                   name,
			Image:                  buildImageName(registry, fn.Image, tag),
			Handler:                fn.Handler,
			Annotations:            fn.Annotations,
			Labels:                 fn.Labels,
			Environment:            environment,
			Constraints:            constraints,
			Limits:                 limits,
			Requests:               requests,
			ReadOnlyRootFilesystem: fn.ReadOnlyRootFilesystem,
		}

		manifest := models.Function{
			TypeMeta: models.TypeMeta{
				APIVersion: api,
				Kind:       resourceKind,
			},
			ObjectMeta: models.ObjectMeta{
				Name:      name,
				Namespace: functionNamespace,
			},
			Spec: spec,
		}

		manifests = append(manifests, manifest)
	}

	//Marshal the object definition to yaml
	out, err := yaml.Marshal(manifests)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}
