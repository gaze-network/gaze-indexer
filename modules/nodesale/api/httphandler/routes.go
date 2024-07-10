package httphandler

import (
	"github.com/gofiber/fiber/v2"
)

func (h *handler) Mount(router fiber.Router) error {
	r := router.Group("/nodesale/v1")

	r.Get("/info", h.infoHandler)
	r.Get("/deploy/:deployId", h.deployHandler)
	r.Get("/nodes", h.nodesHandler)
	r.Get("/events", h.eventsHandler)

	return nil
}
