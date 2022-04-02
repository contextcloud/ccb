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

type Environment map[string]string

type Secret struct {
	Name      string
	Namespace string
	Raw       []byte
}

type FunctionRoute struct {
	Key   string
	Route manifests.FunctionRoute
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
