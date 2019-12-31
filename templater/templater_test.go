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

func Test_Templater(t *testing.T) {
	templater := NewTemplater(
		AddLocationOption("golang-middleware", "github.com/openfaas-incubator/golang-http-template"),
	)
	templater.AddFunction("yes", "golang-es")
	templater.AddFunction("yes2", "golang-middleware")
	templater.AddFunction("yes3", "golang-http-es")

	r, err := templater.Download(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}

	expected := []string{
		"github.com/openfaas-incubator/golang-http-template/template//golang-middleware",
		"github.com/contextgg/openfaas-templates/template//golang-http-es",
		"github.com/contextgg/openfaas-templates/template//golang-es",
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expect %v, Got %v", expected, r)
		return
	}
}
