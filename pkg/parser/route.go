package parser

import (
	"github.com/go-playground/validator/v10"

	"github.com/contextcloud/ccb-cli/pkg/manifests"
)

type Route struct {
	Key      string `validate:"required"`
	FQDN     string `validate:"required"`
	Includes []manifests.RouteInclude
}

func newRoute(key string, raw manifests.Route) (*Route, error) {
	r := &Route{
		Key:      key,
		FQDN:     raw.FQDN,
		Includes: raw.Includes,
	}

	// validate it!
	if err := validator.New().Struct(r); err != nil {
		return nil, err
	}

	return r, nil
}
