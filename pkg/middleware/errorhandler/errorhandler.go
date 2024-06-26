package errorhandler

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gofiber/fiber/v2"
)

// New setup error handler middleware
func New() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		err := ctx.Next()
		if err == nil {
			return nil
		}
		if e := new(errs.PublicError); errors.As(err, &e) {
			return errors.WithStack(ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": e.Message(),
			}))
		}
		if e := new(fiber.Error); errors.As(err, &e) {
			return errors.WithStack(ctx.Status(e.Code).JSON(fiber.Map{
				"error": e.Error(),
			}))
		}
		logger.ErrorContext(ctx.UserContext(), "Something went wrong, api error",
			slogx.String("event", "api_error"),
			slogx.Error(err),
		)
		return errors.WithStack(ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		}))
	}
}
