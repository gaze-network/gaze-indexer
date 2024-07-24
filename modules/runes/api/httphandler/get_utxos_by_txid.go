package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gofiber/fiber/v2"
)

type getUTXOsByTxIdRequest struct {
	txHash    string `params:"txid"`
	outputIdx int32  `query:"outputIndex"`
}

func (r getUTXOsByTxIdRequest) Validate() error {
	var errList []error
	if r.txHash == "" {
		errList = append(errList, errors.New("'txid' is required"))
	}
	if r.outputIdx < 0 {
		errList = append(errList, errors.New("'outputIndex' must be non-negative"))
	}
	return errors.Join(errList...)
}

type getUTXOByTxIdResult struct {
	List []utxoItem `json:"list"`
}

type getUTXOByTxIdResponse = HttpResponse[getUTXOByTxIdResult]

func (h *HttpHandler) GetUTXOsByTxID(ctx *fiber.Ctx) (err error) {
	var req getUTXOsByTxIdRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	resp := getUTXOByTxIdResponse{
		Result: &getUTXOByTxIdResult{},
	}
	return errors.WithStack(ctx.JSON(resp))
}
