package stacktrace

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

// StackTrace is the type of the data for a call stack.
// This mirrors the type of the same name in [github.com/cockroachdb/errors/errbase.StackTrace].
type StackTrace struct {
	PCS    []uintptr
	Frames []Frame
}

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
	pcs := make([]uintptr, 64)
	n := runtime.Callers(2+skip, pcs[:])

	// Expand the pcs slice if there wasn't enough room.
	for n == len(pcs) {
		pcs = make([]uintptr, 2*len(pcs))
		n = runtime.Callers(2+skip, pcs[:])
	}

	// Deallocate the unused space in the slice.
	pcs = pcs[:n:n]

	// Preprocess stack frames.
	frames := make([]Frame, 0, len(pcs))
	callerFrames := runtime.CallersFrames(pcs)
	for frame, more := callerFrames.Next(); more; frame, more = callerFrames.Next() {
		if !strings.HasPrefix(frame.Function, "runtime.") {
			frames = append(frames, Frame{frame})
		}
	}

	return &StackTrace{
		PCS:    pcs,
		Frames: frames[:len(frames):len(frames)],
	}
}

// FramesStrings returns the frames of this stacktrace as slice of strings.
func (s *StackTrace) FramesStrings() []string {
	str := make([]string, len(s.Frames))
	for i, frame := range s.Frames {
		str[i] = frame.String()
	}
	return str
}

// Count reports the total number of frames in this stacktrace.
func (s *StackTrace) Count() int {
	return len(s.Frames)
}

// Format formats the stack of Frames according to the fmt.Formatter interface.
func (s *StackTrace) Format(fs fmt.State, verb rune) {
	for i, frame := range s.Frames {
		if i > 0 {
			_, _ = io.WriteString(fs, "\n")
		}
		frame.Format(fs, verb)
	}
}

// String returns a string representation of the stack trace.
func (s *StackTrace) String() string {
	var sb strings.Builder
	for i, frame := range s.Frames {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("[%d] ", len(s.Frames)-i))
		sb.WriteString(frame.String())
	}
	return sb.String()
}
