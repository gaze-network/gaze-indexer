package slogx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gaze-network/indexer-network/pkg/bufferpool"
	"github.com/gaze-network/indexer-network/pkg/logger/stacktrace"
)

// Any returns an slog.Attr for the supplied value.
// See [AnyValue] for how values are treated.
func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// Group returns an slog.Attr for a Group [Value].
// The first argument is the key; the remaining arguments
// are converted to Attrs as in [Logger.Log].
//
// Use Group to collect several key-value pairs under a single
// key on a log line, or as the result of LogValue
// in order to log a single value as multiple Attrs.
func Group(key string, args ...any) slog.Attr {
	return slog.Group(key, args...)
}

// Error returns an slog.Attr for an error value.
func Error(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.Any(ErrorKey, err)
}

// String returns an slog.Attr for a string value.
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// func Stringp(key string, value *string) slog.Attr {}

// Stringer returns an slog.Attr for a fmt.Stringer value.
func Stringer(key string, value fmt.Stringer) slog.Attr {
	return slog.String(key, value.String())
}

// Int64 returns an slog.Attr for an int64.
func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

// Int32 converts an int32 to an int64 and returns
func Int32(key string, value int32) slog.Attr {
	return Int64(key, int64(value))
}

// Int16 converts an int16 to an int64 and returns
func Int16(key string, value int16) slog.Attr {
	return Int64(key, int64(value))
}

// Int8 converts an int8 to an int64 and returns
func Int8(key string, value int8) slog.Attr {
	return Int64(key, int64(value))
}

// Int converts an int to an int64 and returns
// an slog.Attr with that value.
func Int(key string, value int) slog.Attr {
	return Int64(key, int64(value))
}

// Uint64 returns an slog.Attr for a uint64.
func Uint64(key string, v uint64) slog.Attr {
	return slog.Uint64(key, v)
}

// Uint32 converts a uint32 to a uint64 and returns
func Uint32(key string, v uint32) slog.Attr {
	return Uint64(key, uint64(v))
}

// Uint16 converts a uint16 to a uint64 and returns
func Uint16(key string, v uint16) slog.Attr {
	return Uint64(key, uint64(v))
}

// Uint8 converts a uint8 to a uint64 and returns
func Uint8(key string, v uint8) slog.Attr {
	return Uint64(key, uint64(v))
}

// Uint converts a uint to a uint64 and returns
func Uint(key string, v uint) slog.Attr {
	return Uint64(key, uint64(v))
}

// Uintptr returns an slog.Attr for a uintptr.
func Uintptr(key string, v uintptr) slog.Attr {
	return Uint64(key, uint64(v))
}

// Float64 returns an slog.Attr for a floating-point number.
func Float64(key string, v float64) slog.Attr {
	return slog.Float64(key, v)
}

// Float32 converts a float32 to a float64 and returns
// an slog.Attr with that value.
func Float32(key string, v float32) slog.Attr {
	return Float64(key, float64(v))
}

// Bool returns an slog.Attr for a bool.
func Bool(key string, v bool) slog.Attr {
	return slog.Bool(key, v)
}

// Time returns an slog.Attr for a [time.Time].
// It discards the monotonic portion.
func Time(key string, v time.Time) slog.Attr {
	return slog.Time(key, v)
}

// Duration returns an slog.Attr for a [time.Duration].
func Duration(key string, v time.Duration) slog.Attr {
	return slog.Duration(key, v)
}

// Binary returns an slog.Attr for a binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
// use ByteString.
func Binary(key string, v []byte) slog.Attr {
	return slog.String(key, base64.StdEncoding.EncodeToString(v))
}

// ByteString returns an slog.Attr for a UTF-8 encoded byte string.
//
// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
// Binary.
func ByteString(key string, v []byte) slog.Attr {
	return slog.String(key, string(v))
}

// Reflect returns an slog.Attr for an arbitrary object.
// It uses an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into an slog.Attr, but it's relatively slow and
// allocation-heavy. Any is always a better choice.
func Reflect(key string, v interface{}) slog.Attr {
	buff := bufferpool.Get()
	defer buff.Free()
	enc := json.NewEncoder(buff)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	buff.TrimNewline()
	return slog.String(key, buff.String())
}

// Stack returns an slog.Attr for the current stack trace.
// Keep in mind that taking a stacktrace is eager and
// expensive (relatively speaking); this function both makes an allocation and
// takes about two microseconds.
func Stack(key string) slog.Attr {
	return StackSkip(key, 1)
}

// StackSkip returns an slog.Attr for the stack trace similarly to Stack,
// but also skips the given number of frames from the top of the stacktrace.
func StackSkip(key string, skip int) slog.Attr {
	return slog.Any(key, stacktrace.Capture(skip+1).TraceFramesStrings())
}
