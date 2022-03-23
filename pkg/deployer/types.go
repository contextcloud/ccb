package deployer

import "github.com/contextcloud/ccb-cli/pkg/manifests"

type Probe struct {
	Enabled             bool
	Path                string
	Port                string
	InitialDelaySeconds int
	TimeoutSeconds      int
	PeriodSeconds       int
}

// Environment represents external file for environment data
type Environment map[string]string

type Secret struct {
	Name      string
	Namespace string
	Raw       []byte
}

// KubeSecret for parsing secret files
type KubeSecret struct {
	Metadata *KubeMetadata `yaml:"metadata"`
}

// KubeMetadata basic metadata for a K8 manifest
type KubeMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Function struct {
	Key      string
	Name     string
	Version  string
	Env      map[string]string
	Envs     []string
	Secrets  []string
	Limits   *manifests.FunctionResources
	Requests *manifests.FunctionResources
}
