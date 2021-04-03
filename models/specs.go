package models

// TypeMeta for k8
type TypeMeta struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
}

// ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.
type ObjectMeta struct {
	Name      string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

// Function describes an OpenFaaS function
type Function struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata,omitempty"`
	Spec       FunctionSpec `yaml:"spec"`
}

// FunctionResources is used to set CPU and memory limits and requests
type FunctionResources struct {
	Memory string `yaml:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty"`
}

// FunctionSpec is the spec for a Function resource
type FunctionSpec struct {
	Name                   string             `yaml:"name"`
	Image                  string             `yaml:"image"`
	Replicas               *int               `yaml:"replicas,omitempty"`
	Annotations            *map[string]string `yaml:"annotations,omitempty"`
	Labels                 *map[string]string `yaml:"labels,omitempty"`
	Environment            *map[string]string `yaml:"environment,omitempty"`
	Constraints            []string           `yaml:"constraints,omitempty"`
	Secrets                []string           `yaml:"secrets,omitempty"`
	Limits                 *FunctionResources `yaml:"limits,omitempty"`
	Requests               *FunctionResources `yaml:"requests,omitempty"`
	ReadOnlyRootFilesystem *bool              `yaml:"readOnlyRootFilesystem,omitempty"`
}

// KubeMetadata basic metadata for a K8 manifest
type KubeMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// KubeSecret for parsing secret files
type KubeSecret struct {
	Kind     string                 `yaml:"kind"`
	Metadata *KubeMetadata          `yaml:"metadata"`
	Data     map[string]interface{} `yaml:"data"`
	Spec     *SealedSecretSpec      `yaml:"spec"`
}

// SealedSecretSpec for parsing sealedSecrets encrypted data
type SealedSecretSpec struct {
	EncryptedData map[string]interface{} `yaml:"encryptedData"`
}

// ProjectedSecret for create a project secret volume
type ProjectedSecret struct {
	Name  string
	Items []string
}
