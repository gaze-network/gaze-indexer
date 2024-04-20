package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gofiber/fiber/v2"
)

type getCurrentBlockResult struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
}

type getCurrentBlockResponse = HttpResponse[getCurrentBlockResult]

func (h *HttpHandler) GetCurrentBlock(ctx *fiber.Ctx) (err error) {
	blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
	if err != nil {
		return errors.Wrap(err, "error during GetLatestBlock")
	}

	resp := getCurrentBlockResponse{
		Result: &getCurrentBlockResult{
			Hash:   blockHeader.Hash.String(),
			Height: blockHeader.Height,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
