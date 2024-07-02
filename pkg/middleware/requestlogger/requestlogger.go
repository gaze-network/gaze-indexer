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
	WithRequestHeader    bool     `env:"REQUEST_HEADER" envDefault:"false" mapstructure:"request_header"`
	WithRequestQuery     bool     `env:"REQUEST_QUERY" envDefault:"false" mapstructure:"request_query"`
	Disable              bool     `env:"DISABLE" envDefault:"false" mapstructure:"disable"` // Disable logger level `INFO`
	HiddenRequestHeaders []string `env:"HIDDEN_REQUEST_HEADERS" mapstructure:"hidden_request_headers"`
}

// New setup request context and information
func New(config Config) fiber.Handler {
	hiddenRequestHeaders := make(map[string]struct{}, len(config.HiddenRequestHeaders))
	for _, header := range config.HiddenRequestHeaders {
		hiddenRequestHeaders[strings.TrimSpace(strings.ToLower(header))] = struct{}{}
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
		if config.WithRequestQuery {
			requestAttributes = append(requestAttributes, slog.String("query", string(c.Request().URI().QueryString())))
		}

		// request headers
		if config.WithRequestHeader {
			kv := []any{}

			for k, v := range c.GetReqHeaders() {
				if _, found := hiddenRequestHeaders[strings.ToLower(k)]; found {
					continue
				}
				kv = append(kv, slog.Any(k, v))
			}

			requestAttributes = append(requestAttributes, slog.Group("header", kv...))
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
