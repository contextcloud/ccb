package templater

import (
	"context"
	"reflect"
	"testing"
)

func Test_Can_Download(t *testing.T) {
	err := download("github.com/contextgg/openfaas-templates/openfaas//golang-http-es", "golang-http-es")
	if err != nil {
		t.Error(err)
	}
}

func Test_Templater_Download(t *testing.T) {
	templater := NewTemplater(
		AddLocationOption("golang-middleware", "https://github.com/openfaas-incubator/golang-http-template"),
	)
	templater.AddFunction("yes2", "template", "golang-middleware")
	templater.AddFunction("yes3", "openfaas", "golang-http-es")

	r, err := templater.Download(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}

	expected := 2
	if len(r) != expected {
		t.Errorf("Expect %v, Got %v", expected, r)
		return
	}
}

func Test_Pack(t *testing.T) {
	err := pack("golang-http-es", "example", nil)
	if err != nil {
		t.Error(err)
	}
}

func Test_Templater_Pack(t *testing.T) {
	templater := NewTemplater()
	templater.AddFunction("example", "openfaas", "golang-http-es")

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
