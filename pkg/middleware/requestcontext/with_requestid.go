package requestcontext

import (
	"context"

	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	fiberutils "github.com/gofiber/fiber/v2/utils"
)

type requestIdKey struct{}

// GetRequestId get requestId from context. If not found, return empty string
//
// Warning: Request context should be setup before using this function
func GetRequestId(ctx context.Context) string {
	if id, ok := ctx.Value(requestIdKey{}).(string); ok {
		return id
	}
	return ""
}

func WithRequestId() Option {
	return func(ctx context.Context, c *fiber.Ctx) (context.Context, error) {
		// Try to get id from fiber context.
		requestId, ok := c.Locals(requestid.ConfigDefault.ContextKey).(string)
		if !ok || requestId == "" {
			// Try to get id from request, else we generate one
			requestId = c.Get(requestid.ConfigDefault.Header, fiberutils.UUID())

			// Set new id to response header
			c.Set(requestid.ConfigDefault.Header, requestId)

			// Add the request ID to locals (fasthttp UserValue storage)
			c.Locals(requestid.ConfigDefault.ContextKey, requestId)
		}

		// Add the request ID to context
		ctx = context.WithValue(ctx, requestIdKey{}, requestId)

		// Add the requuest ID to context logger
		ctx = logger.WithContext(ctx, "requestId", requestId)

		return ctx, nil
	}
}
