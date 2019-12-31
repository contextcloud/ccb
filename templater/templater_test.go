package templater

import (
	"context"
	"reflect"
	"testing"
)

func Test_Can_Download(t *testing.T) {
	err := download("github.com/contextgg/openfaas-templates/template//golang-es", "golang-es")
	if err != nil {
		t.Error(err)
	}
}

func Test_Templater_Download(t *testing.T) {
	templater := NewTemplater(
		AddLocationOption("golang-middleware", "https://github.com/openfaas-incubator/golang-http-template"),
	)
	templater.AddFunction("yes", "golang-es")
	templater.AddFunction("yes2", "golang-middleware")
	templater.AddFunction("yes3", "golang-http-es")

	r, err := templater.Download(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}

	expected := 3
	if len(r) != expected {
		t.Errorf("Expect %v, Got %v", expected, r)
		return
	}
}

func Test_Pack(t *testing.T) {
	err := pack("golang-es", "example", nil)
	if err != nil {
		t.Error(err)
	}
}

func Test_Templater_Pack(t *testing.T) {
	templater := NewTemplater()
	templater.AddFunction("example", "golang-es")

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
