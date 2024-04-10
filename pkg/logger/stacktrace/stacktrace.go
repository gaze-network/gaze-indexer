package stacktrace

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/cockroachdb/errors/errbase"
)

// StackTrace is the type of the data for a call stack.
// This mirrors the type of the same name in [github.com/cockroachdb/errors/errbase.StackTrace].
type StackTrace errbase.StackTrace

// Caller captures a stack trace of the specified depth, skipping
// the provided number of frames. skip=0 identifies the caller of Caller.
//
// Alias of [Capture]
func Caller(skip int) *StackTrace {
	return Capture(1 + skip)
}

// Capture captures a stack trace of the specified depth, skipping
// the provided number of frames. skip=0 identifies the caller of Capture.
func Capture(skip int) *StackTrace {
	const numFrames = 32
	var pcs [numFrames]uintptr
	n := runtime.Callers(2+skip, pcs[:])
	f := make([]errbase.StackFrame, n)
	for i := 0; i < len(f); i++ {
		f[i] = errbase.StackFrame(pcs[i])
	}
	return (*StackTrace)(&f)
}

// TraceFrames returns the trace line frames of the stack trace.
func (s StackTrace) TraceFrames() []TraceFrame {
	return TraceLines(s)
}

func (s StackTrace) TraceFramesStrings() []string {
	traceLines := s.TraceFrames()
	t := make([]string, len(traceLines))
	for i, tl := range traceLines {
		t[i] = tl.String()
	}
	return t
}

// Format formats the stack of Frames according to the fmt.Formatter interface.
func (s StackTrace) Format(fs fmt.State, verb rune) {
	tracelines := s.TraceFrames()
	for i, tl := range tracelines {
		if i > 0 {
			_, _ = io.WriteString(fs, "\n")
		}
		tl.Format(fs, verb)
	}
}

// String returns a string representation of the stack trace.
func (s StackTrace) String() string {
	var sb strings.Builder
	tracelines := s.TraceFrames()
	for i, tl := range tracelines {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("[%d] ", i+1))
		sb.WriteString(tl.String())
	}
	return sb.String()
}
