package logger

import (
	"context"
	"log/slog"
)

type (
	handleFunc func(context.Context, slog.Record) error
	middleware func(handleFunc) handleFunc
)

type multiHandlers struct {
	h           slog.Handler
	middlewares []middleware
}

func newChainHandlers(handler slog.Handler, middlewares ...middleware) *multiHandlers {
	return &multiHandlers{
		h:           handler,
		middlewares: middlewares,
	}
}

func (c *multiHandlers) Enabled(ctx context.Context, lvl slog.Level) bool {
	return c.h.Enabled(ctx, lvl)
}

func (c *multiHandlers) Handle(ctx context.Context, rec slog.Record) error {
	h := c.h.Handle
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h(ctx, rec)
}

func (c *multiHandlers) WithGroup(group string) slog.Handler {
	return &multiHandlers{
		middlewares: c.middlewares,
		h:           c.h.WithGroup(group),
	}
}

func (c *multiHandlers) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &multiHandlers{
		middlewares: c.middlewares,
		h:           c.h.WithAttrs(attrs),
	}
}
