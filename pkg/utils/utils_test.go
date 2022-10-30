package utils

import "testing"

func TestYaml(t *testing.T) {
	p, err := YamlFile("", "../deployer/example/.secrets/assets.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(p)
}
