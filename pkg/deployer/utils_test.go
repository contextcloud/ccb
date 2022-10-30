package deployer

import "testing"

func Test_LoadSecret(t *testing.T) {
	secret, err := LoadSecret("./example/.secrets/assets.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	if secret.Name != "assets" {
		t.Error("Invalid name")
		return
	}

	if secret.Namespace != "" {
		t.Error("Invalid namespace")
		return
	}

	t.Log(secret)
}

func Test_LoadEnv(t *testing.T) {
	env, err := LoadEnv("./example/.env/common.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	if env["demo"] == nil || *env["demo"] != "yes" {
		t.Error("Invalid env")
		return
	}

	t.Log(env)
}
