package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gofiber/fiber/v2"
)

type infoResponse struct {
	IndexedBlockHeight int64  `json:"indexedBlockHeight"`
	IndexedBlockHash   string `json:"indexedBlockHash"`
}

func (h *handler) infoHandler(ctx *fiber.Ctx) error {
	block, err := h.nodeSaleDg.GetLastProcessedBlock(ctx.UserContext())
	if err != nil {
		return errors.Wrap(err, "Cannot get last processed block")
	}
	err = ctx.JSON(infoResponse{
		IndexedBlockHeight: block.BlockHeight,
		IndexedBlockHash:   block.BlockHash,
	})
	if err != nil {
		return errors.Wrap(err, "Go fiber cannot parse JSON")
	}
	return nil
}
