package requestlogger

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/middleware/requestcontext"
	"github.com/gofiber/fiber/v2"
)

type Config struct {
	AllRequestHeaders    bool     `env:"REQUEST_HEADER" envDefault:"false" mapstructure:"request_header"` // Log all request headers
	AllRequestQueries    bool     `env:"REQUEST_QUERY" envDefault:"false" mapstructure:"request_query"`   // Log all request queries
	Disable              bool     `env:"DISABLE" envDefault:"false" mapstructure:"disable"`               // Disable logger level `INFO`
	HiddenRequestHeaders []string `env:"HIDDEN_REQUEST_HEADERS" mapstructure:"hidden_request_headers"`    // Hide specific headers from log
	WithRequestHeaders   []string `env:"WITH_REQUEST_HEADERS" mapstructure:"with_request_headers"`        // Add specific headers to log (higher priority than `HiddenRequestHeaders`)                                  // Additional fields to log
}

// New setup request context and information
func New(config Config) fiber.Handler {
	hiddenRequestHeaders := make(map[string]struct{}, len(config.HiddenRequestHeaders))
	for _, header := range config.HiddenRequestHeaders {
		hiddenRequestHeaders[strings.TrimSpace(strings.ToLower(header))] = struct{}{}
	}
	withRequestHeaders := make(map[string]struct{}, len(config.WithRequestHeaders))
	for _, header := range config.WithRequestHeaders {
		withRequestHeaders[strings.TrimSpace(strings.ToLower(header))] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Continue stack
		err := c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Response().StatusCode()

		baseAttrs := []slog.Attr{
			slog.String("event", "api_request"),
			slog.Int64("latency", latency.Milliseconds()),
			slog.String("latencyHuman", latency.String()),
		}

		// prep request attributes
		requestAttributes := []slog.Attr{
			slog.Time("time", start),
			slog.String("method", c.Method()),
			slog.String("host", c.Hostname()),
			slog.String("path", c.Path()),
			slog.String("route", c.Route().Path),
			slog.String("ip", requestcontext.GetClientIP(c.UserContext())),
			slog.String("remoteIP", c.Context().RemoteIP().String()),
			slog.Any("x-forwarded-for", c.IPs()),
			slog.String("user-agent", string(c.Context().UserAgent())),
			slog.Any("params", c.AllParams()),
			slog.Any("query", c.Queries()),
			slog.Int("length", len((c.Body()))),
		}

		// prep response attributes
		responseAttributes := []slog.Attr{
			slog.Time("time", end),
			slog.Int("status", status),
			slog.Int("length", len(c.Response().Body())),
		}

		// request query
		if config.AllRequestQueries {
			requestAttributes = append(requestAttributes, slog.String("queries", string(c.Request().URI().QueryString())))
		}

		// request headers
		if config.AllRequestHeaders || len(config.WithRequestHeaders) > 0 {
			kv := []any{}

			for k, v := range c.GetReqHeaders() {
				h := strings.ToLower(k)

				// add headers for WithRequestHeaders
				if _, found := withRequestHeaders[h]; found {
					goto add
				}

				// skip hidden headers
				if _, found := hiddenRequestHeaders[h]; found {
					continue
				}

			add:
				val := any(v)
				if len(v) == 1 {
					val = v[0]
				}
				kv = append(kv, slog.Any(k, val))
			}

			requestAttributes = append(requestAttributes, slog.Group("headers", kv...))
		}

		level := slog.LevelInfo
		if err != nil || status >= http.StatusInternalServerError {
			level = slog.LevelError

			// error attributes
			logErr := err
			if logErr == nil {
				logErr = fiber.NewError(status)
			}
			baseAttrs = append(baseAttrs, slog.Any("error", logErr))
		}

		if config.Disable && level == slog.LevelInfo {
			return errors.WithStack(err)
		}

		logger.LogAttrs(c.UserContext(), level, "Request Completed", append([]slog.Attr{
			{
				Key:   "request",
				Value: slog.GroupValue(requestAttributes...),
			},
			{
				Key:   "response",
				Value: slog.GroupValue(responseAttributes...),
			},
		}, baseAttrs...)...,
		)

		return errors.WithStack(err)
	}
}
