package deployer

import "github.com/contextcloud/ccb/pkg/manifests"

type Resources struct {
	Requests *ResourceValues `yaml:"requests"`
	Limits   *ResourceValues `yaml:"limits"`
}

type ResourceValues struct {
	Memory string `yaml:"memory"`
	CPU    string `yaml:"cpu"`
}

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
	Kind        string
	Name        string
	Namespace   string
	Raw         []byte
	SecretNames []string
}

type FunctionRoute struct {
	Key   string
	Route manifests.FunctionRoute
}

// KubeSecret for parsing secret files
type KubeSecret struct {
	Kind     string        `yaml:"kind"`
	Metadata *KubeMetadata `yaml:"metadata"`
}

// KubeMetadata basic metadata for a K8 manifest
type KubeMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type KubeSopSecret struct {
	Kind     string             `yaml:"kind"`
	Metadata *KubeMetadata      `yaml:"metadata"`
	Spec     *KubeSopSecretSpec `yaml:"spec"`
}

type KubeSopSecretSpec struct {
	SecretTemplates []struct {
		Name string `yaml:"name"`
	} `yaml:"secretTemplates"`
}
