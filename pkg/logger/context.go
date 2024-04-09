package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

type loggerKey struct{}

// FromContext returns the logger from the context. If no logger is found, a new
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return logger.With()
	}

	if log, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return log
	}

	return logger.With()
}

// NewContext returns a new context with logger attached.
func NewContext(ctx context.Context, log *slog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, loggerKey{}, log)
}

// WithContext returns a new context with given logger attributes.
func WithContext(ctx context.Context, args ...any) context.Context {
	return NewContext(ctx, FromContext(ctx).With(args...))
}

// WithGroupContext returns a new context with given group.
func WithGroupContext(ctx context.Context, group string) context.Context {
	return NewContext(ctx, FromContext(ctx).WithGroup(group))
}

// DebugContext logs at [LevelDebug] from logger in the given context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	log(ctx, FromContext(ctx), slog.LevelDebug, msg, args...)
}

// InfoContext logs at [LevelInfo] from logger in the given context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	log(ctx, FromContext(ctx), slog.LevelInfo, msg, args...)
}

// WarnContext logs at [LevelWarn] from logger in the given context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	log(ctx, FromContext(ctx), slog.LevelWarn, msg, args...)
}

// ErrorContext logs at [LevelError] from logger in the given context.
func ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	log(ctx, FromContext(ctx), slog.LevelError, msg, append(args, slogx.Error(err))...)
}

// PanicContext logs at [LevelPanic] and then panics from logger in the given context.
func PanicContext(ctx context.Context, msg string, args ...any) {
	log(ctx, FromContext(ctx), LevelPanic, msg, args...)
	panic(msg)
}

// FatalContext logs at [LevelFatal] and then [os.Exit](1) from logger in the given context.
func FatalContext(ctx context.Context, msg string, args ...any) {
	log(ctx, FromContext(ctx), LevelFatal, msg, args...)
	os.Exit(1)
}

// LogContext logs at the given level from logger in the given context.
func LogContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	log(ctx, FromContext(ctx), level, msg, args...)
}
