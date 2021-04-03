package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/Masterminds/sprig"

	"github.com/contextcloud/ccb-cli/cmd/templates"
	"github.com/contextcloud/ccb-cli/models"
	"github.com/contextcloud/ccb-cli/spec"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
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

func findFile(fileName string) (io.ReadCloser, error) {
	paths := []string{
		"./" + fileName + ".yaml",
		"./" + fileName + ".yml",
		"./" + fileName,
		fileName + ".yaml",
		fileName + ".yml",
		fileName,
	}

	for _, p := range paths {
		file, err := os.Open(p)
		if err == nil {
			return file, nil
		}
	}

	return nil, os.ErrNotExist
}

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	// Add some extra functionality
	extra := template.FuncMap{
		"toYaml": toYAML,
		// "fromYaml": fromYAML,
	}
	for k, v := range extra {
		f[k] = v
	}
	return f
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func getSecrets(namespace string, items []string) ([]*models.ProjectedSecret, error) {
	var secrets []*models.ProjectedSecret

	for _, s := range items {
		file, err := findFile(s)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var secret models.KubeSecret
		if err := yaml.NewDecoder(file).Decode(&secret); err != nil {
			return nil, fmt.Errorf("error decoding secret %q: %w", s, err)
		}

		if secret.Metadata == nil {
			return nil, fmt.Errorf("error decoding secret %q: %w", s, ErrNoMetadata)
		}

		// Validate namespace!
		if namespace != secret.Metadata.Namespace {
			return nil, fmt.Errorf("error decoding secret %q: %w", s, ErrInvalidNamespace)
		}

		s := &models.ProjectedSecret{
			Name:  secret.Metadata.Name,
			Items: []string{},
		}

		switch secret.Kind {
		case "Secret":
			// save the keys
			for key := range secret.Data {
				s.Items = append(s.Items, key)
			}
		case "SealedSecret":
			if secret.Spec == nil {
				return nil, fmt.Errorf("error decoding secret %q: %w", s, ErrNoSpec)
			}
			// save the keys
			for key := range secret.Spec.EncryptedData {
				s.Items = append(s.Items, key)
			}
		default:
			return nil, fmt.Errorf("Error decoding secret: %q: %w", s, ErrInvalidKind)
		}

		secrets = append(secrets, s)
	}

	return secrets, nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	parsedServices, err := spec.ParseYAMLFile(yamlFile, regex, filter, envsubst)
	if err != nil {
		return err
	}

	box := templates.NewBox()

	var manifests []string
	for name, fn := range parsedServices.Functions {
		g := NewGen(box, name, &fn)

		all, err := g.Execute()
		if err != nil {
			return err
		}
		manifests = append(manifests, all...)
	}

	var all string
	//Marshal the object definitions to yaml
	for i, manifest := range manifests {
		if i > 0 {
			all += "\n"
		}
		all += "---\n" + manifest
	}
	fmt.Println(all)
	return nil
}

type probe struct {
	Enabled             bool
	Path                string
	Port                string
	InitialDelaySeconds int
	TimeoutSeconds      int
	PeriodSeconds       int
}

type gen struct {
	box  *rice.Box
	name string
	fn   *spec.Function
}

func (g *gen) cloud() ([]string, error) {
	data := make(map[string]interface{})

	namespace := g.fn.Namespace
	if len(namespace) == 0 {
		namespace = functionNamespace
	}

	//read environment variables from the file
	fileEnvironment, err := readFiles(g.fn.EnvironmentFile)
	if err != nil {
		return nil, err
	}
	//combine all environment variables
	allEnvironment, envErr := compileEnvironment([]string{}, g.fn.Environment, fileEnvironment)
	if envErr != nil {
		return nil, envErr
	}

	var limits map[string]string
	if g.fn.Limits != nil {
		limits = make(map[string]string)
		if len(g.fn.Limits.Memory) > 0 {
			limits["memory"] = g.fn.Limits.Memory
		}
		if len(g.fn.Limits.CPU) > 0 {
			limits["cpu"] = g.fn.Limits.CPU
		}
	}
	var requests map[string]string
	if g.fn.Requests != nil {
		requests = make(map[string]string)
		if len(g.fn.Requests.Memory) > 0 {
			requests["memory"] = g.fn.Requests.Memory
		}
		if len(g.fn.Limits.CPU) > 0 {
			requests["cpu"] = g.fn.Requests.CPU
		}
	}
	var constraints []string
	if g.fn.Constraints != nil {
		constraints = *g.fn.Constraints
	}

	var environment *map[string]string
	if len(allEnvironment) > 0 {
		environment = &allEnvironment
	}

	var readOnlyRoot *bool = nil
	if g.fn.ReadOnlyRootFilesystem {
		readOnlyRoot = &g.fn.ReadOnlyRootFilesystem
	}

	livenessPort := "health"
	if g.fn.Liveness.Port != "" {
		livenessPort = g.fn.Liveness.Port
	}
	var livenessProbe *probe = &probe{
		Enabled:             !g.fn.Liveness.Disabled,
		Path:                "/live",
		Port:                livenessPort,
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
	}

	readinessPort := "health"
	if g.fn.Readiness.Port != "" {
		readinessPort = g.fn.Readiness.Port
	}
	var readinessProbe *probe = &probe{
		Enabled:             !g.fn.Readiness.Disabled,
		Path:                "/ready",
		Port:                readinessPort,
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
	}

	secrets, err := getSecrets(namespace, g.fn.Secrets)
	if err != nil {
		return nil, err
	}

	data["Name"] = g.name
	data["Namespace"] = namespace
	data["Image"] = buildImageName(registry, g.name, tag)
	data["Annotations"] = g.fn.Annotations
	data["Labels"] = g.fn.Labels
	data["Environment"] = environment
	data["Constraints"] = constraints
	data["Limits"] = limits
	data["Requests"] = requests
	data["ReadOnlyRootFilesystem"] = readOnlyRoot
	data["LivenessProbe"] = livenessProbe
	data["ReadinessProbe"] = readinessProbe
	data["Secrets"] = secrets

	if len(g.fn.SqlProxy) > 0 {
		data["SqlProxy"] = g.fn.SqlProxy
	}

	var all []string
	gen := func(path string, info os.FileInfo, err error) error {
		// Guard
		if err != nil {
			return err
		}

		// If it's a dir skip
		if info.IsDir() {
			return nil
		}

		tmplString, err := g.box.String(path)
		if err != nil {
			return err
		}

		// parse the file
		tmpl, err := template.New(path).Funcs(funcMap()).Parse(tmplString)
		if err != nil {
			return err
		}

		var tpl bytes.Buffer
		if err := tmpl.Execute(&tpl, data); err != nil {
			return err
		}

		all = append(all, tpl.String())
		return nil
	}

	name := "deployment"
	if err := g.box.Walk(name, gen); err != nil {
		return nil, err
	}

	return all, nil
}

func (g *gen) openfaas() ([]string, error) {
	//read environment variables from the file
	fileEnvironment, err := readFiles(g.fn.EnvironmentFile)
	if err != nil {
		return nil, err
	}

	//combine all environment variables
	allEnvironment, envErr := compileEnvironment([]string{}, g.fn.Environment, fileEnvironment)
	if envErr != nil {
		return nil, envErr
	}

	var limits *models.FunctionResources
	if g.fn.Limits != nil {
		limits = &models.FunctionResources{
			Memory: g.fn.Limits.Memory,
			CPU:    g.fn.Limits.CPU,
		}
	}
	var requests *models.FunctionResources
	if g.fn.Requests != nil {
		requests = &models.FunctionResources{
			Memory: g.fn.Requests.Memory,
			CPU:    g.fn.Requests.CPU,
		}
	}
	var constraints []string
	if g.fn.Constraints != nil {
		constraints = *g.fn.Constraints
	}

	var environment *map[string]string
	if len(allEnvironment) > 0 {
		environment = &allEnvironment
	}

	var readOnlyRoot *bool = nil
	if g.fn.ReadOnlyRootFilesystem {
		readOnlyRoot = &g.fn.ReadOnlyRootFilesystem
	}

	spec := models.FunctionSpec{
		Name:                   g.name,
		Image:                  buildImageName(registry, g.name, tag),
		Annotations:            g.fn.Annotations,
		Labels:                 g.fn.Labels,
		Secrets:                g.fn.Secrets,
		Environment:            environment,
		Constraints:            constraints,
		Limits:                 limits,
		Requests:               requests,
		ReadOnlyRootFilesystem: readOnlyRoot,
	}

	namespace := g.fn.Namespace
	if len(namespace) == 0 {
		namespace = functionNamespace
	}
	manifest := models.Function{
		TypeMeta: models.TypeMeta{
			APIVersion: api,
			Kind:       resourceKind,
		},
		ObjectMeta: models.ObjectMeta{
			Name:      spec.Name,
			Namespace: namespace,
		},
		Spec: spec,
	}

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return []string{string(out)}, nil
}

func (g *gen) Execute() ([]string, error) {
	switch strings.ToLower(g.fn.Engine) {
	case "cloud":
		return g.cloud()
	default:
		return g.openfaas()
	}
}

func NewGen(box *rice.Box, name string, fn *spec.Function) *gen {
	return &gen{
		box:  box,
		name: name,
		fn:   fn,
	}
}
