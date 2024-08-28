package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/constants"
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
		if !errors.Is(err, errs.NotFound) {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeader = constants.StartingBlockHeader[h.network]
	}

	resp := getCurrentBlockResponse{
		Result: &getCurrentBlockResult{
			Hash:   blockHeader.Hash.String(),
			Height: blockHeader.Height,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
