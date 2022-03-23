package deployer

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

func ImageName(registry, imageName, tag string) string {
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

func LoadEnv(filename string) (Environment, error) {
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var env Environment
	if err := yaml.Unmarshal(out, &env); err != nil {
		return nil, err
	}

	return env, nil
}

func LoadSecret(filename string) (*Secret, error) {
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var secret KubeSecret
	if err := yaml.Unmarshal(out, &secret); err != nil {
		return nil, err
	}

	if secret.Metadata == nil {
		return nil, ErrNoMetadata
	}

	return &Secret{
		Name:      secret.Metadata.Name,
		Namespace: secret.Metadata.Namespace,
		Raw:       out,
	}, nil
}
