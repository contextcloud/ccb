package deployer

import (
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
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

	secretNames := []string{secret.Metadata.Name}

	if strings.ToLower(secret.Kind) == "sopssecret" {
		var sopsSecret KubeSopSecret
		if err := yaml.Unmarshal(out, &sopsSecret); err != nil {
			return nil, err
		}
		secretNames = make([]string, len(sopsSecret.Spec.SecretTemplates))
		for i, s := range sopsSecret.Spec.SecretTemplates {
			secretNames[i] = s.Name
		}
	}

	return &Secret{
		Kind:        secret.Kind,
		Name:        secret.Metadata.Name,
		Namespace:   secret.Metadata.Namespace,
		Raw:         out,
		SecretNames: secretNames,
	}, nil
}

func GetFuncMaps(namespacePrefix string, routePrefix string) template.FuncMap {
	fm := sprig.TxtFuncMap()
	fm["toYaml"] = func(v interface{}) string {
		data, err := yaml.Marshal(v)
		if err != nil {
			// Swallow errors inside of a template.
			return ""
		}
		return strings.TrimSuffix(string(data), "\n")
	}
	fm["namespace"] = func(v interface{}) string {
		ns, ok := v.(string)
		if !ok {
			return ""
		}
		return namespacePrefix + ns
	}
	fm["route"] = func(v interface{}) string {
		ns, ok := v.(string)
		if !ok {
			return ""
		}
		return routePrefix + ns
	}
	return fm
}
