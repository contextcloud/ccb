package parser

import (
	"fmt"

	"github.com/contextcloud/ccb/pkg/manifests"
	"github.com/ryanuber/go-glob"
)

type Stack interface {
	GetRoutes(filters ...string) ([]*Route, error)
	GetFunctions(filters ...string) ([]*Function, error)
}

type stack struct {
	raw *manifests.Stack
}

func (s *stack) isMatch(name string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if glob.Glob(filter, name) {
			return true
		}
	}

	return false
}

func (s *stack) GetRoutes(filters ...string) ([]*Route, error) {
	var routes []*Route

	// filter using input
	for k, raw := range s.raw.Routes {
		route, err := newRoute(k, raw)
		if err != nil {
			return nil, err
		}

		if !s.isMatch(route.Key, filters) {
			continue
		}

		routes = append(routes, route)
	}

	return routes, nil

}

func (s *stack) GetFunctions(filters ...string) ([]*Function, error) {
	var fns []*Function

	// filter using input
	for k, raw := range s.raw.Functions {
		fn, err := newFunction(k, raw)
		if err != nil {
			return nil, err
		}

		if !s.isMatch(fn.Key, filters) {
			continue
		}

		fns = append(fns, fn)
	}

	return fns, nil
}

func NewStack(raw *manifests.Stack) (Stack, error) {
	// validate version.
	if !isValidSchemaVersion(raw.Provider.Version) {
		return nil, fmt.Errorf("%s are the only valid versions for the stack file - found: %s", ValidSchemaVersions, raw.Provider.Version)
	}

	return &stack{raw}, nil
}
