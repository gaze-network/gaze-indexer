package logger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/stacktrace"
)

func middlewareError() middleware {
	return func(next handleFunc) handleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			rec.Attrs(func(attr slog.Attr) bool {
				if attr.Key == slogx.ErrorKey || attr.Key == "err" {
					err := attr.Value.Any()
					if err, ok := err.(error); ok && err != nil {
						rec.AddAttrs(slog.String("error_verbose", fmt.Sprintf("%+v", err)))
						if st, ok := stacktrace.ParseErrStackTrace(err); ok {
							rec.AddAttrs(slog.Any("stack_trace", st.FramesStrings()))
						}
					}
				}
				return false
			})
			return next(ctx, rec)
		}
	}
}
