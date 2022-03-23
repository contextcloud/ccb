package print

import (
	"os"
)

// Env is the environment of a command
type Env struct {
	Out Out
	Err Out
}

func NewEnv() *Env {
	return &Env{
		Out: out{Writer: os.Stdout},
		Err: out{Writer: os.Stderr},
	}
}
