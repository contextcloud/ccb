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

	manager := NewManager("../../example", "default")
	fns, err := stack.GetFunctions()

	all := make([]*Function, len(fns))
	for i, fn := range fns {
		f := &Function{
			Key:      fn.Key,
			Name:     fn.Name,
			Version:  fn.Version,
			Env:      fn.Env,
			Envs:     fn.Envs,
			Secrets:  fn.Secrets,
			Limits:   fn.Limits,
			Requests: fn.Requests,
			Routes:   fn.Routes,
		}
		all[i] = f
	}

	manifests, err := manager.GenerateFunctions("", "latest", all)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(manifests.merged())
}
