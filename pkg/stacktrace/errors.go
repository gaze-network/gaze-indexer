package stacktrace

import (
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors/errbase"
	"github.com/samber/lo"
)

// ErrorStackTrace is a pair of an error and its stack trace.
type ErrorStackTrace struct {
	Cause      error
	StackTrace *StackTrace
}

func (s ErrorStackTrace) String() string {
	return fmt.Sprintf("%s %v", s.Cause.Error(), s.StackTrace.FramesStrings())
}

func (s ErrorStackTrace) Error() string {
	return s.Cause.Error()
}

// nolint: errcheck
func (s ErrorStackTrace) Format(f fmt.State, verb rune) {
	fmt.Fprintf(f, "%s %v", s.Cause.Error(), s.StackTrace.FramesStrings())
}

// ErrorStackTraces is a list of error stack traces.
type ErrorStackTraces []ErrorStackTrace

func (s ErrorStackTraces) String() string {
	var sb strings.Builder
	for i, errSt := range s {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("[%d] ", i+1))
		sb.WriteString(errSt.String())
	}
	return sb.String()
}

// nolint: errcheck
func (s ErrorStackTraces) Format(f fmt.State, verb rune) {
	for i, errSt := range s {
		if i > 0 {
			io.WriteString(f, "\n")
		}
		io.WriteString(f, fmt.Sprintf("[%d] %s", i+1, errSt.String()))
	}
}

// ExtractErrorStackTraces extracts the stack traces from the provided error and its causes.
// Sorted from oldest to newest.
func ExtractErrorStackTraces(err error) ErrorStackTraces {
	result := ErrorStackTraces{}

	for err != nil {
		causeErr := errbase.UnwrapOnce(err)
		if errStack, ok := err.(errbase.StackTraceProvider); ok {
			pcs := pkgErrStackTaceToPCs(errStack.StackTrace())
			result = append(result, ErrorStackTrace{
				Cause:      err,
				StackTrace: ParsePCS(pcs),
			})
		}
		err = causeErr
	}

	// reverse the order (oldest first)
	result = lo.Reverse(result)

	return result
}

// convert type of [github.com/cockroachdb/errors/errbase.StackTrace] to a slice of PCs.
func pkgErrStackTaceToPCs(stacktrace errbase.StackTrace) []uintptr {
	pcs := make([]uintptr, len(stacktrace))
	for i, frame := range stacktrace {
		pcs[i] = uintptr(frame)
	}
	return pcs
}
