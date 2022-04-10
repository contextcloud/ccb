package templater

import (
	"context"
	"reflect"
	"testing"
)

func Test_Templater_Pack(t *testing.T) {
	ctx := context.TODO()

	templater := NewTemplater("./example")
	templater.AddFunction("assets", "golang")

	_, err := templater.Download(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	r, err := templater.Pack(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	expected := []string{
		"assets",
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expect %v, Got %v", expected, r)
		return
	}
}
func Test_Templater_Tar(t *testing.T) {
	ctx := context.TODO()

	templater := NewTemplater("./example")
	templater.AddFunction("assets", "golang")

	_, err := templater.Download(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	r, err := templater.Tar(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	expected := []string{
		"assets",
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expect %v, Got %v", expected, r)
		return
	}
}
