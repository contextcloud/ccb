package print

import (
	"os"
)

type Logger interface {
	Out() Log
	Err() Log
}

type logger struct {
	out Log
	err Log
}

func (l *logger) Out() Log {
	return l.out
}

func (l *logger) Err() Log {
	return l.err
}

func NewConsoleLogger() Logger {
	return &logger{
		out: NewLog(os.Stdout),
		err: NewLog(os.Stderr),
	}
}
