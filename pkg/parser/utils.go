package parser

import (
	"os"

	"github.com/drone/envsubst"
)

func substituteEnvironment(data []byte) ([]byte, error) {
	ret, err := envsubst.Parse(string(data))
	if err != nil {
		return nil, err
	}

	res, resErr := ret.Execute(func(input string) string {
		if val, ok := os.LookupEnv(input); ok {
			return val
		}
		return ""
	})

	return []byte(res), resErr
}
