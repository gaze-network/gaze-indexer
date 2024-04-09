package logger

import (
	"fmt"
	"log/slog"
)

const (
	LevelCritical = slog.Level(12)
	LevelPanic    = slog.Level(14)
	LevelFatal    = slog.Level(16)
)

func levelAttrReplacer(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && attr.Key == "level" {
		str := func(base string, val slog.Level) string {
			if val == 0 {
				return base
			}
			return fmt.Sprintf("%s%+d", base, val)
		}

		if l, ok := attr.Value.Any().(slog.Level); ok {
			switch {
			case l < LevelCritical:
				return attr
			case l < LevelPanic:
				return slog.Attr{Key: attr.Key, Value: slog.StringValue(str("CRITICAL", l-LevelCritical))}
			case l < LevelFatal:
				return slog.Attr{Key: attr.Key, Value: slog.StringValue(str("PANIC", l-LevelPanic))}
			default:
				return slog.Attr{Key: attr.Key, Value: slog.StringValue(str("FATAL", l-LevelFatal))}
			}
		}
	}
	return attr
}
