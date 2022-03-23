package manifests

// Stack is a stack of functions
type Stack struct {
	Provider  Provider            `yaml:"provider,omitempty"`
	Functions map[string]Function `yaml:"functions,omitempty"`
}

// Provider for the FaaS set of functions.
type Provider struct {
	Version string `yaml:"version,omitempty"`
}

// FunctionResources is used to set CPU and memory limits and requests
type FunctionResources struct {
	Memory string `yaml:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty"`
}

// Function as deployed or built on FaaS
type Function struct {
	Name         string             `yaml:"name"`
	Version      string             `yaml:"version"`
	Template     string             `yaml:"template"`
	BuildOptions []string           `yaml:"build_options"`
	BuildArgs    map[string]string  `yaml:"build_args"`
	Env          map[string]string  `yaml:"env"`
	Secrets      []string           `yaml:"secrets"`
	Envs         []string           `yaml:"envs"`
	Labels       *map[string]string `yaml:"labels"`
	Annotations  *map[string]string `yaml:"annotations"`
	Replicas     *int               `yaml:"replicas"`
	Limits       *FunctionResources `yaml:"limits"`
	Requests     *FunctionResources `yaml:"requests"`
}
