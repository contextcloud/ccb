package manifests

// FunctionResources is used to set CPU and memory limits and requests
type FunctionResources struct {
	Memory string `yaml:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty"`
}

// RouterHeader is used to set the header for the router
type RouteHeader struct {
	Name     string `yaml:"name"`
	Operator string `yaml:"op"`
	Value    string `yaml:"value"`
}

// Stack is a stack of functions
type Stack struct {
	Provider  Provider            `yaml:"provider,omitempty"`
	Functions map[string]Function `yaml:"functions,omitempty"`
	Routes    map[string]Route    `yaml:"routes,omitempty"`
}

// Provider for the FaaS set of functions.
type Provider struct {
	Version string `yaml:"version,omitempty"`
}

// FunctionRoute is a route to a function
type FunctionRoute struct {
	Name     string        `yaml:"name,omitempty"`
	Prefix   string        `yaml:"prefix,omitempty"`
	Headers  []RouteHeader `yaml:"headers,omitempty"`
	Redirect string        `yaml:"redirect,omitempty"`
}

// Function as deployed or built
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
	Routes       []FunctionRoute    `yaml:"routes,omitempty"`
}

// RouteInclude is a route to a namespace
type RouteInclude struct {
	Namespace string        `yaml:"namespace,omitempty"`
	Name      string        `yaml:"name,omitempty"`
	Prefix    string        `yaml:"prefix,omitempty"`
	Headers   []RouteHeader `yaml:"headers,omitempty"`
}

type Route struct {
	Name     string          `yaml:"name,omitempty"`
	FQDN     string          `yaml:"fqdn,omitempty"`
	Routes   []FunctionRoute `yaml:"routes,omitempty"`
	Includes []RouteInclude  `yaml:"includes,omitempty"`
}
