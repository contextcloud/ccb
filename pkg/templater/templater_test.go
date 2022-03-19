package templater

import (
	"context"
	"reflect"
	"testing"
)

func Test_Can_Download(t *testing.T) {
	err := download("github.com/contextcloud/templates//golang", "golang")
	if err != nil {
		t.Error(err)
	}
}

func Test_Pack(t *testing.T) {
	err := pack("golang", "example", nil)
	if err != nil {
		t.Error(err)
	}
}

func Test_Templater_Pack(t *testing.T) {
	templater := NewTemplater()
	templater.AddFunction("example", "golang")

	_, err := templater.Download(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}
	r, err := templater.Pack(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}

	expected := []string{
		"example",
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expect %v, Got %v", expected, r)
		return
	}
}
