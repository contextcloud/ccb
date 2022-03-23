package parser

import (
	"github.com/go-playground/validator/v10"

	"github.com/contextcloud/ccb-cli/pkg/manifests"
)

type Function struct {
	Key      string `validate:"required"`
	Name     string `validate:"required"`
	Version  string `validate:"required"`
	Template string `validate:"required"`

	BuildArgs map[string]string

	Env      map[string]string
	Envs     []string
	Secrets  []string
	Limits   *manifests.FunctionResources
	Requests *manifests.FunctionResources
}

func newFunction(key string, raw manifests.Function) (*Function, error) {
	n := raw.Name
	if len(n) == 0 {
		n = key
	}

	fn := &Function{
		Key:       key,
		Name:      n,
		Version:   raw.Version,
		Template:  raw.Template,
		BuildArgs: raw.BuildArgs,
		Env:       raw.Env,
		Envs:      raw.Envs,
		Secrets:   raw.Secrets,
		Limits:    raw.Limits,
		Requests:  raw.Requests,
	}

	// validate it!
	if err := validator.New().Struct(fn); err != nil {
		return nil, err
	}

	return fn, nil
}
