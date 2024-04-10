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
						rec.AddAttrs(slog.String(slogx.ErrorVerboseKey, fmt.Sprintf("%+v", err)))
						if st, ok := stacktrace.ParseErrStackTrace(err); ok {
							rec.AddAttrs(slog.Any(slogx.ErrorStackTraceKey, st.FramesStrings()))
						}
					}
				}
				return false
			})
			return next(ctx, rec)
		}
	}
}

func errorAttrReplacer(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && (attr.Key == slogx.ErrorKey || attr.Key == "err") {
		if err, ok := attr.Value.Any().(error); ok {
			if err != nil {
				return slog.Attr{Key: slogx.ErrorKey, Value: slog.StringValue(err.Error())}
			}
			return slog.Attr{Key: slogx.ErrorKey, Value: slog.StringValue("null")}
		}
	}
	return attr
}
