package logger

import (
	"log/slog"
	"os"
)

func NewGCPHandler(opts *slog.HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     opts.Level,
		ReplaceAttr: attrReplacerChain(
			GCPAttrReplacer,
			opts.ReplaceAttr,
		),
	})
}

// GCPAttrReplacer replaces the default attribute keys with the GCP logging attribute keys.
func GCPAttrReplacer(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case MessageKey:
		attr.Key = "message"
	case SourceKey:
		attr.Key = "logging.googleapis.com/sourceLocation"
	case LevelKey:
		attr.Key = "severity"
		lvl, ok := attr.Value.Any().(slog.Level)
		if ok {
			attr.Value = slog.StringValue(gcpSeverityMapping(lvl))
		}
	}
	return attr
}

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
func gcpSeverityMapping(lvl slog.Level) string {
	switch {
	case lvl < slog.LevelInfo:
		return "DEBUG"
	case lvl < slog.LevelWarn:
		return "INFO"
	case lvl < slog.LevelError:
		return "WARNING"
	case lvl < LevelCritical:
		return "ERROR"
	case lvl < LevelPanic:
		return "CRITICAL"
	case lvl < LevelFatal:
		return "ALERT"
	default:
		return "EMERGENCY"
	}
}
