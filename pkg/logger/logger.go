// nolint: sloglint
package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
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

// With returns a Logger that includes the given attributes
// in each output operation. Arguments are converted to
// attributes as if by [Logger.Log].
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

// WithGroup returns a Logger that starts a group, if name is non-empty.
// The keys of all attributes added to the Logger will be qualified by the given
// name. (How that qualification happens depends on the [Handler.WithGroup]
// method of the Logger's Handler.)
//
// If name is empty, WithGroup returns the receiver.
func WithGroup(group string) *slog.Logger {
	return logger.WithGroup(group)
}

// Debug logs at [LevelDebug].
func Debug(msg string, args ...any) {
	log(context.Background(), logger, slog.LevelDebug, msg, args...)
}

// Info logs at [LevelInfo].
func Info(msg string, args ...any) {
	log(context.Background(), logger, slog.LevelInfo, msg, args...)
}

// Warn logs at [LevelWarn].
func Warn(msg string, args ...any) {
	log(context.Background(), logger, slog.LevelWarn, msg, args...)
}

// Error logs at [LevelError] with an error.
func Error(msg string, args ...any) {
	log(context.Background(), logger, slog.LevelError, msg, args...)
}

// Panic logs at [LevelPanic] and then panics.
func Panic(msg string, args ...any) {
	log(context.Background(), logger, LevelPanic, msg, args...)
	panic(msg)
}

// Fatal logs at [LevelFatal] followed by a call to [os.Exit](1).
func Fatal(msg string, args ...any) {
	log(context.Background(), logger, LevelFatal, msg, args...)
	os.Exit(1)
}

// Log emits a log record with the current time and the given level and message.
// The Record's Attrs consist of the Logger's attributes followed by
// the Attrs specified by args.
func Log(level slog.Level, msg string, args ...any) {
	log(context.Background(), logger, level, msg, args...)
}

// LogAttrs is a more efficient version of [Logger.Log] that accepts only Attrs.
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, FromContext(ctx), level, msg, attrs...)
}

// Config is the logger configuration.
type Config struct {
	// Output is the logger output format.
	// Possible values:
	//  - Text (default)
	//  - JSON
	//  - GCP: Output format for Stackdriver Logging/Cloud Logging or others GCP services.
	Output string `mapstructure:"output" env:"OUTPUT" envDefault:"text"`

	// Debug is enabled logger level debug. (default: false)
	Debug bool `mapstructure:"debug" env:"DEBUG" envDefault:"false"`
}

var (
	// Default Attribute Replacers
	defaultAttrReplacers = []func([]string, slog.Attr) slog.Attr{
		levelAttrReplacer,
		errorAttrReplacer,
	}

	// Default Middlewares
	defaultMiddleware = []middleware{}
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
		middlewares = append([]middleware{}, defaultMiddleware...)
	)

	lvl.Set(slog.LevelInfo)
	if cfg.Debug {
		lvl.Set(slog.LevelDebug)
		options.AddSource = true
		middlewares = append(middlewares, middlewareErrorStackTrace())
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

	logger = slog.New(newChainHandlers(handler, middlewares...))
	slog.SetDefault(logger)
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

// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func log(ctx context.Context, l *slog.Logger, level slog.Level, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !l.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = l.Handler().Handle(ctx, r)
}

// logAttrs is like [Logger.log], but for methods that take ...Attr.
func logAttrs(ctx context.Context, l *slog.Logger, level slog.Level, msg string, attrs ...slog.Attr) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !l.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	_ = l.Handler().Handle(ctx, r)
}
