package httphandler

import (
	"github.com/gofiber/fiber/v2"
)

func (h *HttpHandler) Mount(router fiber.Router) error {
	r := router.Group("/v2/runes")

	r.Post("/balances/wallet/batch", h.GetBalancesBatch)
	r.Get("/balances/wallet/:wallet", h.GetBalances)
	r.Get("/transactions", h.GetTransactions)
	r.Get("/holders/:id", h.GetHolders)
	r.Get("/info/:id", h.GetTokenInfo)
	r.Get("/utxos/wallet/:wallet", h.GetUTXOs)
	r.Post("/utxos/output/batch", h.GetUTXOsOutputByLocationBatch)
	r.Get("/utxos/output/:txHash", h.GetUTXOsOutputByLocation)
	r.Get("/block", h.GetCurrentBlock)
	r.Get("/tokens", h.GetTokenList)
	return nil
}
