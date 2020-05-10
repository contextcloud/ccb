package spec

// Services root level YAML file to define FaaS function-set
type Services struct {
	Provider           Provider            `yaml:"provider,omitempty"`
	StackConfiguration StackConfiguration  `yaml:"configuration,omitempty"`
	Functions          map[string]Function `yaml:"functions,omitempty"`
}

// Provider for the FaaS set of functions.
type Provider struct {
	Version string `yaml:"version,omitempty"`
}

// TemplateSource for build templates
type TemplateSource struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source,omitempty"`
}

// StackConfiguration for the overall stack.yml
type StackConfiguration struct {
	TemplateConfigs []TemplateSource `yaml:"templates"`
	// CopyExtraPaths specifies additional paths (relative to the stack file) that will be copied
	// into the functions build context, e.g. specifying `"common"` will look for and copy the
	// "common/" folder of file in the same root as the stack file.  All paths must be contained
	// within the project root defined by the location of the stack file.
	//
	// The yaml uses the shorter name `copy` to make it easier for developers to read and use
	CopyExtraPaths []string `yaml:"copy"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `yaml:"memory"`
	CPU    string `yaml:"cpu"`
}

// Function as deployed or built on FaaS
type Function struct {
	// Name of deployed function
	Name string `yaml:"-"`

	// Engine either openfaas or cloud
	Engine string `yaml:"engine"`

	Language string `yaml:"lang"`

	// Image Docker image name
	Image string `yaml:"image"`

	Environment map[string]string `yaml:"environment"`

	// Secrets list of secrets to be made available to function
	Secrets []string `yaml:"secrets"`

	SkipBuild bool `yaml:"skip_build"`

	Constraints *[]string `yaml:"constraints"`

	// EnvironmentFile is a list of files to import and override environmental variables.
	// These are overriden in order.
	EnvironmentFile []string `yaml:"environment_file"`

	Labels *map[string]string `yaml:"labels"`

	// Limits for function
	Limits *FunctionResources `yaml:"limits"`

	// Requests of resources requested by function
	Requests *FunctionResources `yaml:"requests"`

	// ReadOnlyRootFilesystem is used to set the container filesystem to read-only
	ReadOnlyRootFilesystem bool `yaml:"readonly_root_filesystem"`

	// BuildOptions to determine native packages
	BuildOptions []string `yaml:"build_options"`

	// BuildArgs for the function
	BuildArgs map[string]string `yaml:"build_args"`

	// Annotations
	Annotations *map[string]string `yaml:"annotations"`

	// Namespace of the function
	Namespace string `yaml:"namespace,omitempty"`
}
