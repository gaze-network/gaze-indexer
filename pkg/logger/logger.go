// nolint: sloglint
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

const (
	// DefaultLevel is the default minimum reporting level for the logger
	DefaultLevel = slog.LevelDebug

	// logLevel set `log` output level to `DEBUG`.
	// `log` is allowed for debugging purposes only.
	//
	// NOTE: Please use `slog` for logging instead of `log`, and
	// do not use `log` for production code.
	logLevel = slog.LevelDebug
)

var (
	// minimum reporting level for the logger
	lvl = new(slog.LevelVar)

	// top-level logger
	logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       lvl,
		ReplaceAttr: levelAttrReplacer,
	}))
)

// Set default slog logger
func init() {
	lvl.Set(DefaultLevel)
	slog.SetDefault(logger)
}

// Set `log` output level
func init() {
	slog.SetLogLoggerLevel(logLevel)
}

// SetLevel sets the minimum reporting level for the logger
func SetLevel(level slog.Level) (old slog.Level) {
	old = lvl.Level()
	lvl.Set(level)
	return old
}

// Debug calls [Logger.Debug] on the default logger.
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

// Debug calls [Logger.Debug] on the default logger.
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info calls [Logger.Info] on the default logger.
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Warn calls [Logger.Warn] on the default logger.
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error calls [Logger.Error] on the default logger.
// TODO: support stack trace for error
func Error(msg string, err error, args ...any) {
	logger.Error(msg, append(args, AttrError(err))...)
}

// Panic calls [Logger.Log] with PANIC level on the default logger and then panic.
func Panic(msg string, args ...any) {
	logger.Log(context.Background(), LevelPanic, msg, args...)
	panic(msg)
}

// Log calls [Logger.Log] on the default logger.
func Log(level slog.Level, msg string, args ...any) {
	logger.Log(context.Background(), level, msg, args...)
}

// LogAttrs calls [Logger.LogAttrs] on the default logger.
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	logger.LogAttrs(ctx, level, msg, attrs...)
}

// Config is the logger configuration.
type Config struct {
	// Env is the logger environment.
	//	- PRODUCTION, PROD: use JSON format, log level: INFO
	//	- Default: use Text format, log level: DEBUG
	Env string `env:"ENV,expand" envDefault:"${ENV}"`

	Platform string `env:"PLATFORM" envDefault:"none"`
}

// Init initializes global logger and slog logger with given configuration.
func Init(cfg Config) error {
	replacers := []func([]string, slog.Attr) slog.Attr{}

	// Platform specific attr replacer
	switch strings.ToLower(cfg.Platform) {
	case "gcp":
		replacers = append(replacers, GCPAttrReplacer)
	}

	// Default attr replacer
	replacers = append(replacers,
		levelAttrReplacer,
	)

	var (
		handler slog.Handler
		level   = new(slog.LevelVar)
		options = &slog.HandlerOptions{
			AddSource:   true,
			Level:       level,
			ReplaceAttr: attrReplacerChain(replacers...),
		}
	)

	switch strings.ToLower(cfg.Env) {
	case "production", "prod":
		level.Set(slog.LevelInfo)
		handler = slog.NewJSONHandler(os.Stdout, options)
	default:
		level.Set(DefaultLevel)
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	logger = slog.New(newChainHandlers(handler, middlewareError()))

	lvl = level
	slog.SetDefault(logger)

	logger.Info("logger initialized", slog.String("environment", cfg.Env))
	return nil
}

// attrReplacerChain returns a function that applies a chain of replacers to an attribute.
func attrReplacerChain(replacers ...func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, attr slog.Attr) slog.Attr {
		for _, replacer := range replacers {
			attr = replacer(groups, attr)
		}
		return attr
	}
}
