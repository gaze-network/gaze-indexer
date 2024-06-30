package httphandler

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gofiber/fiber/v2"
)

// TODO: use modules/brc20/constants.go
var startingBlockHeader = map[common.Network]types.BlockHeader{
	common.NetworkMainnet: {
		Height: 767429,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000000002b35aef66eb15cd2b232a800f75a2f25cedca4cfe52c4")),
	},
	common.NetworkTestnet: {
		Height: 2413342,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000022e97030b143af785de812f836dd0651b6ac2b7dd9e90dc9abf9")),
	},
}

type getCurrentBlockResult struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
}

type getCurrentBlockResponse = common.HttpResponse[getCurrentBlockResult]

func (h *HttpHandler) GetCurrentBlock(ctx *fiber.Ctx) (err error) {
	blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
	if err != nil {
		if !errors.Is(err, errs.NotFound) {
			return errors.Wrap(err, "error during get latest block")
		}
		blockHeader = startingBlockHeader[h.network]
	}

	resp := getCurrentBlockResponse{
		Result: &getCurrentBlockResult{
			Hash:   blockHeader.Hash.String(),
			Height: blockHeader.Height,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
