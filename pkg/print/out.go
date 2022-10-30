package print

import (
	"fmt"
	"io"
)

type Log interface {
	Printf(format string, a ...interface{})
	Print(a ...interface{})
	Println(a ...interface{})
}

type log struct {
	io.Writer
}

func (o log) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(o, format, a...)
}

func (o log) Print(a ...interface{}) {
	_, _ = fmt.Fprint(o, a...)
}

func (o log) Println(a ...interface{}) {
	_, _ = fmt.Fprintln(o, a...)
}

func NewLog(w io.Writer) Log {
	return log{w}
}
