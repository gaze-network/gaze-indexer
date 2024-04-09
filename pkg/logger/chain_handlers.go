package logger

import (
	"context"
	"log/slog"
)

type (
	handleFunc func(context.Context, slog.Record) error
	middleware func(handleFunc) handleFunc
)

type chainHandlers struct {
	h           slog.Handler
	middlewares []middleware
}

func newChainHandlers(handler slog.Handler, middlewares ...middleware) *chainHandlers {
	return &chainHandlers{
		h:           handler,
		middlewares: middlewares,
	}
}

func (c *chainHandlers) Enabled(ctx context.Context, lvl slog.Level) bool {
	return c.h.Enabled(ctx, lvl)
}

func (c *chainHandlers) Handle(ctx context.Context, rec slog.Record) error {
	h := c.h.Handle
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h(ctx, rec)
}

func (c *chainHandlers) WithGroup(group string) slog.Handler {
	return &chainHandlers{
		middlewares: c.middlewares,
		h:           c.h.WithGroup(group),
	}
}

func (c *chainHandlers) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &chainHandlers{
		middlewares: c.middlewares,
		h:           c.h.WithAttrs(attrs),
	}
}
