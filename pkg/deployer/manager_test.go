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

	manager := NewManager("../../example", "default", "v1")
	fns, err := stack.GetFunctions()
	if err != nil {
		t.Error(err)
		return
	}

	manifests, err := manager.GenerateFunctions("", "latest", fns)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(manifests.merged())
}

func Test_Routes(t *testing.T) {
	stackFile := path.Join("../../example", "stack.yml")

	stack, err := parser.LoadStack(stackFile)
	if err != nil {
		t.Error(err)
		return
	}

	manager := NewManager("../../example", "default", "v1")
	rts, err := stack.GetRoutes()
	if err != nil {
		t.Error(err)
		return
	}

	manifests, err := manager.GenerateRoutes(rts)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(manifests.merged())
}
