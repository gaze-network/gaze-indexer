package httphandler

import (
	"github.com/gofiber/fiber/v2"
)

func (h *HttpHandler) Mount(router fiber.Router) error {
	r := router.Group("/v2/runes")

	r.Post("/balances/wallet/batch", h.GetBalancesByAddressBatch)
	r.Get("/balances/wallet/:wallet", h.GetBalancesByAddress)
	r.Get("/transactions", h.GetTransactions)
	r.Get("/holders/:id", h.GetHolders)
	r.Get("/info/:id", h.GetTokenInfo)
	r.Get("/utxos/wallet/:wallet", h.GetUTXOsByAddress)
	return nil
}
