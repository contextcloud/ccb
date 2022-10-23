package parser

import (
	"github.com/go-playground/validator/v10"

	"github.com/contextcloud/ccb/pkg/manifests"
)

type Function struct {
	manifests.Function
	Key string `validate:"required"`
}

func newFunction(key string, raw manifests.Function) (*Function, error) {
	n := raw.Name
	if len(n) == 0 {
		n = key
	}
	raw.Name = n

	fn := &Function{
		Function: raw,
		Key:      key,
	}

	// validate it!
	if err := validator.New().Struct(fn); err != nil {
		return nil, err
	}

	return fn, nil
}
