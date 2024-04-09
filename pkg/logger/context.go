package logger

import (
	"context"
	"log/slog"
	"os"
)

// DebugContext logs at [LevelDebug] with the given context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	log(ctx, logger, slog.LevelDebug, msg, args...)
}

// InfoContext logs at [LevelInfo] with the given context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	log(ctx, logger, slog.LevelInfo, msg, args...)
}

// WarnContext logs at [LevelWarn] with the given context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	log(ctx, logger, slog.LevelWarn, msg, args...)
}

// ErrorContext logs at [LevelError] with an error and the given context.
func ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	log(ctx, logger, slog.LevelError, msg, append(args, AttrError(err))...)
}

// PanicContext logs at [LevelPanic] and then panics with the given context.
func PanicContext(ctx context.Context, msg string, args ...any) {
	log(ctx, logger, LevelPanic, msg, args...)
	panic(msg)
}

// FatalContext logs at [LevelFatal] and then [os.Exit](1) with the given context.
func FatalContext(ctx context.Context, msg string, args ...any) {
	log(ctx, logger, LevelFatal, msg, args...)
	os.Exit(1)
}
