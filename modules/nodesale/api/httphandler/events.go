package httphandler

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/nodesale/protobuf"
	"github.com/gofiber/fiber/v2"
)

type eventRequest struct {
	WalletAddress string `query:"walletAddress"`
}

type eventResposne struct {
	TxHash         string          `json:"txHash"`
	BlockHeight    int64           `json:"blockHeight"`
	TxIndex        int32           `json:"txIndex"`
	WalletAddress  string          `json:"walletAddress"`
	Action         string          `json:"action"`
	ParsedMessage  json.RawMessage `json:"parsedMessage"`
	BlockTimestamp time.Time       `json:"blockTimestamp"`
	BlockHash      string          `json:"blockHash"`
}

func (h *handler) eventsHandler(ctx *fiber.Ctx) error {
	var request eventRequest
	err := ctx.QueryParser(&request)
	if err != nil {
		return errors.Wrap(err, "cannot parse query")
	}

	events, err := h.nodeSaleDg.GetEventsByWallet(ctx.UserContext(), request.WalletAddress)
	if err != nil {
		return errors.Wrap(err, "Can't get events from db")
	}

	responses := make([]eventResposne, len(events))
	for i, event := range events {
		responses[i].TxHash = event.TxHash
		responses[i].BlockHeight = event.BlockHeight
		responses[i].TxIndex = event.TxIndex
		responses[i].WalletAddress = event.WalletAddress
		responses[i].Action = protobuf.Action_name[event.Action]
		responses[i].ParsedMessage = event.ParsedMessage
		responses[i].BlockTimestamp = event.BlockTimestamp
		responses[i].BlockHash = event.BlockHash
	}

	err = ctx.JSON(responses)
	if err != nil {
		return errors.Wrap(err, "Go fiber cannot parse JSON")
	}
	return nil
}
