package stacktrace

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

type TraceFrame struct {
	PC       uintptr
	Function string
	File     string
	Line     int
}

func TraceLines(s StackTrace) []TraceFrame {
	traceLines := make([]TraceFrame, 0, len(s))

	// Iterate in reverse to skip uninteresting, consecutive runtime frames at
	// the bottom of the trace.
	skipping := true
	for i := len(s) - 1; i >= 0; i-- {
		// Adapted from errors.Frame.MarshalText(), but avoiding repeated
		// calls to FuncForPC and FileLine.
		pc := uintptr(s[i]) - 1
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			traceLines = append(traceLines, TraceFrame{pc, "unknown", "", 0})
			skipping = false
			continue
		}

		name := fn.Name()
		if skipping && strings.HasPrefix(name, "runtime.") {
			continue
		} else {
			skipping = false
		}

		filename, line := fn.FileLine(pc)
		traceLines = append(traceLines, TraceFrame{pc, name, filename, line})
	}

	return traceLines[:len(traceLines):len(traceLines)]
}

func (f TraceFrame) String() string {
	return fmt.Sprintf("%s %s:%d", f.Function, f.File, f.Line)
}

// nolint: errcheck
func (f TraceFrame) Format(fs fmt.State, verb rune) {
	io.WriteString(fs, f.Function)
	io.WriteString(fs, "\n\t")
	io.WriteString(fs, f.File)
	io.WriteString(fs, ":")
	io.WriteString(fs, fmt.Sprint(f.Line))
}
