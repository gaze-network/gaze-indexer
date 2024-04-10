package stacktrace

import (
	"fmt"
	"io"
	"runtime"
)

type Frame struct {
	runtime.Frame
}

func (f Frame) String() string {
	return fmt.Sprintf("%s %s:%d", f.Function, f.File, f.Line)
}

// nolint: errcheck
func (f Frame) Format(fs fmt.State, verb rune) {
	io.WriteString(fs, f.Function)
	io.WriteString(fs, "\n\t")
	io.WriteString(fs, f.File)
	io.WriteString(fs, ":")
	io.WriteString(fs, fmt.Sprint(f.Line))
}
