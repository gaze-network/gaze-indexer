package httphandler

import (
	"github.com/gofiber/fiber/v2"
)

func (h *HttpHandler) Mount(router fiber.Router) error {
	r := router.Group("/v2/runes")

	r.Get("/balances/wallet/:wallet", h.GetBalancesByAddress)
	r.Get("/transactions", h.GetTransactions)
	r.Get("/info/:id", h.GetTokenInfo)
	return nil
}
