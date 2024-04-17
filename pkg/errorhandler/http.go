package errorhandler

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gofiber/fiber/v2"
)

func NewHTTPErrorHandler() func(ctx *fiber.Ctx, err error) error {
	return func(ctx *fiber.Ctx, err error) error {
		if e := new(errs.PublicError); errors.As(err, e) {
			return errors.WithStack(ctx.Status(http.StatusBadRequest).JSON(map[string]any{
				"error": e.Message(),
			}))
		}
		if e := new(fiber.Error); errors.As(err, &e) {
			return errors.WithStack(ctx.Status(e.Code).SendString(e.Error()))
		}

		logger.ErrorContext(ctx.UserContext(), "unhandled error", slogx.Error(err))
		return errors.WithStack(ctx.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error": "Internal Server Error",
		}))
	}
}
