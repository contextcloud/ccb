package parser

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/contextcloud/ccb-cli/pkg/manifests"
)

type Route struct {
	manifests.Route
	Key string `validate:"required"`
}

func newRoute(key string, raw manifests.Route) (*Route, error) {
	r := &Route{
		Route: raw,
		Key:   key,
	}

	fmt.Println(raw)

	// validate it!
	if err := validator.New().Struct(r); err != nil {
		return nil, err
	}

	return r, nil
}
