//go:generate rice embed-go

package templates

import (
	"strings"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

func NewBox() *rice.Box {
	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},
	}
	return conf.MustFindBox(".")
}

func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func GetFuncMaps(prefix string) template.FuncMap {
	fm := sprig.TxtFuncMap()
	fm["toYaml"] = toYAML
	fm["namespace"] = func(v interface{}) string {
		ns, ok := v.(string)
		if !ok {
			return ""
		}
		return prefix + ns
	}
	return fm
}
