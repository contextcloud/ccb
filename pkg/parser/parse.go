package parser

import (
	"io/ioutil"

	"github.com/contextcloud/ccb/pkg/manifests"
	"gopkg.in/yaml.v2"
)

// LoadStack from a remote url or locally
func LoadStack(yamlFile string) (Stack, error) {
	fileData, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}

	substData, err := substituteEnvironment(fileData)
	if err != nil {
		return nil, err
	}

	var raw manifests.Stack
	if err := yaml.Unmarshal(substData, &raw); err != nil {
		return nil, err
	}

	return NewStack(&raw)
}
