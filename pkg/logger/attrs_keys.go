package logger

import "log/slog"

// Keys for log attributes.
const (
	TimeKey            = slog.TimeKey
	LevelKey           = slog.LevelKey
	MessageKey         = slog.MessageKey
	SourceKey          = slog.SourceKey
	ErrorKey           = "error"
	ErrorVerboseKey    = "error_verbose"
	ErrorStackTraceKey = "error_stacktrace"
)
