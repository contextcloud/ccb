package deployer

import (
	"path"
	"testing"

	"github.com/contextcloud/ccb-cli/pkg/parser"
)

func Test_Build(t *testing.T) {
	stackFile := path.Join("../../example", "stack.yml")

	stack, err := parser.LoadStack(stackFile)
	if err != nil {
		t.Error(err)
		return
	}

	manager := NewManager("../../example", "default", "", "latest")
	fns, err := stack.GetFunctions()
	for _, fn := range fns {
		f := &Function{
			Key:      fn.Key,
			Name:     fn.Name,
			Version:  fn.Version,
			Env:      fn.Env,
			Envs:     fn.Envs,
			Secrets:  fn.Secrets,
			Limits:   fn.Limits,
			Requests: fn.Requests,
		}
		manager.AddFunction(f)
	}

	manifests, err := manager.Generate()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(manifests.merged())
}
