package logger

import (
	"log/slog"
)

func durationToMsAttrReplacer(groups []string, attr slog.Attr) slog.Attr {
	if attr.Value.Kind() == slog.KindDuration {
		return slog.Attr{
			Key:   attr.Key,
			Value: slog.Int64Value(attr.Value.Duration().Milliseconds()),
		}
	}
	return attr
}
