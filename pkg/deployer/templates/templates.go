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

func FuncMap() template.FuncMap {
	extra := template.FuncMap{
		"toYaml": toYAML,
	}
	for k, v := range sprig.TxtFuncMap() {
		extra[k] = v
	}
	return extra
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}
