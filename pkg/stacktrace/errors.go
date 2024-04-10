package stacktrace

import "github.com/cockroachdb/errors/errbase"

// ParseErrStackTrace attempts to parse the stack trace from the provided error.
//
// Supported error types are those that implement the [github.com/cockroachdb/errors/errbase.StackTraceProvider] interface.
func ParseErrStackTrace(err error) (*StackTrace, bool) {
	if errStack, ok := err.(errbase.StackTraceProvider); ok {
		stackTrace := errStack.StackTrace()
		pcs := make([]uintptr, len(stackTrace))
		for i, frame := range stackTrace {
			pcs[i] = uintptr(frame)
		}
		return ParsePCS(pcs), true
	}
	return nil, false
}
