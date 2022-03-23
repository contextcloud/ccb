package print

import (
	"fmt"
	"io"
)

type Out interface {
	Printf(format string, a ...interface{})
	Print(a ...interface{})
	Println(a ...interface{})
}

type out struct {
	io.Writer
}

func (o out) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(o, format, a...)
}

func (o out) Print(a ...interface{}) {
	_, _ = fmt.Fprint(o, a...)
}

func (o out) Println(a ...interface{}) {
	_, _ = fmt.Fprintln(o, a...)
}

func NewOut(w io.Writer) Out {
	return out{w}
}
