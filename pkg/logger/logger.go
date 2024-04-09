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

// DebugContext calls [Logger.DebugContext] on the default logger.
func DebugContext(ctx context.Context, msg string, args ...any) {
	logger.DebugContext(ctx, msg, args...)
}

// Info calls [Logger.Info] on the default logger.
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// InfoContext calls [Logger.InfoContext] on the default logger.
func InfoContext(ctx context.Context, msg string, args ...any) {
	logger.InfoContext(ctx, msg, args...)
}

// Warn calls [Logger.Warn] on the default logger.
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// WarnContext calls [Logger.WarnContext] on the default logger.
func WarnContext(ctx context.Context, msg string, args ...any) {
	logger.WarnContext(ctx, msg, args...)
}

// Error calls [Logger.Error] on the default logger.
func Error(msg string, err error, args ...any) {
	logger.Error(msg, append(args, AttrError(err))...)
}

// ErrorContext calls [Logger.ErrorContext] on the default logger.
func ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	logger.ErrorContext(ctx, msg, append(args, AttrError(err))...)
}

// Panic calls [Logger.Log] with PANIC level on the default logger and then panic.
func Panic(msg string, args ...any) {
	logger.Log(context.Background(), LevelPanic, msg, args...)
	panic(msg)
}

// PanicContext calls [Logger.Log] with PANIC level on the default logger and then panic.
func PanicContext(ctx context.Context, msg string, args ...any) {
	logger.Log(ctx, LevelPanic, msg, args...)
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
	// Output is the logger output format.
	// Possible values:
	//  - Text (default)
	//  - JSON
	//  - GCP: Output format for Stackdriver Logging/Cloud Logging or others GCP services.
	Output string `env:"OUTPUT" envDefault:"TEXT"`

	// Debug is enabled logger level debug. (default: false)
	Debug bool `env:"DEBUG" envDefault:"false"`
}

var (
	// Default Attribute Replacers
	defaultAttrReplacers = []func([]string, slog.Attr) slog.Attr{
		levelAttrReplacer,
	}

	// Default Middlewares
	defaultMiddleware = []middleware{
		middlewareError(),
	}
)

// Init initializes global logger and slog logger with given configuration.
func Init(cfg Config) error {
	var (
		handler slog.Handler
		options = &slog.HandlerOptions{
			AddSource:   false,
			Level:       lvl,
			ReplaceAttr: attrReplacerChain(defaultAttrReplacers...),
		}
	)

	lvl.Set(slog.LevelInfo)
	if cfg.Debug {
		lvl.Set(slog.LevelDebug)
		options.AddSource = true
	}

	switch strings.ToLower(cfg.Output) {
	case "json":
		lvl.Set(slog.LevelInfo)
		handler = slog.NewJSONHandler(os.Stdout, options)
	case "gcp":
		handler = NewGCPHandler(options)
	default:
		handler = slog.NewTextHandler(os.Stdout, options)
	}

	logger = slog.New(newChainHandlers(handler, defaultMiddleware...))
	slog.SetDefault(logger)

	logger.Info("logger initialized", slog.String("log_output", cfg.Output))
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
