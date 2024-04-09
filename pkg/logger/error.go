package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/cockroachdb/errors/errbase"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

func middlewareError() middleware {
	return func(next handleFunc) handleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			rec.Attrs(func(attr slog.Attr) bool {
				if attr.Key == slogx.ErrorKey || attr.Key == "err" {
					err := attr.Value.Any()
					if err, ok := err.(error); ok && err != nil {
						rec.AddAttrs(slog.String("error_verbose", fmt.Sprintf("%+v", err)))
						if x, ok := err.(errbase.StackTraceProvider); ok {
							rec.AddAttrs(slog.Any("stack_trace", traceLines(x.StackTrace())))
						}
					}
				}
				return false
			})

			return next(ctx, rec)
		}
	}
}

func traceLines(frames errbase.StackTrace) []string {
	traceLines := make([]string, 0, len(frames))

	// Iterate in reverse to skip uninteresting, consecutive runtime frames at
	// the bottom of the trace.
	skipping := true
	for i := len(frames) - 1; i >= 0; i-- {
		// Adapted from errors.Frame.MarshalText(), but avoiding repeated
		// calls to FuncForPC and FileLine.
		pc := uintptr(frames[i]) - 1
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			traceLines = append(traceLines, "unknown")
			skipping = false
			continue
		}

		name := fn.Name()
		if skipping && strings.HasPrefix(name, "runtime.") {
			continue
		} else {
			skipping = false
		}

		filename, lineNr := fn.FileLine(pc)
		traceLines = append(traceLines, fmt.Sprintf("%s %s:%d", name, filename, lineNr))
	}

	return traceLines[:len(traceLines):len(traceLines)]
}
