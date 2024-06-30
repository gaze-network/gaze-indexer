package requestcontext

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
}

type Option func(ctx context.Context, c *fiber.Ctx) (context.Context, error)

func New(opts ...Option) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error
		ctx := c.UserContext()
		for i, opt := range opts {
			ctx, err = opt(ctx, c)
			if err != nil {
				rErr := requestcontextError{}
				if errors.As(err, &rErr) {
					return c.Status(rErr.status).JSON(Response{Error: rErr.message})
				}

				logger.ErrorContext(ctx, "failed to extract request context",
					err,
					slog.String("event", "requestcontext/error"),
					slog.String("module", "requestcontext"),
					slog.Int("optionIndex", i),
				)
				return c.Status(http.StatusInternalServerError).JSON(Response{Error: "internal server error"})
			}
		}
		c.SetUserContext(ctx)
		return c.Next()
	}
}
