package adapters

import (
	"context"
	"log/slog"
)

// ErrorHandler receives adapter error to be processed.
type ErrorHandler interface {
	Handle(ctx context.Context, err error)
}

// SlogErrorHandler is a adapter error handler which logs an error.
type SlogErrorHandler struct {
	logger *slog.Logger
}

func NewSlogErrorHandler(logger *slog.Logger) *SlogErrorHandler {
	return &SlogErrorHandler{
		logger: logger,
	}
}

func (h *SlogErrorHandler) Handle(ctx context.Context, err error) {
	h.logger.Error("error", "err", err)
}

// NoOpErrorHandler is an adapter error handler which does nothing.
type NoOpErrorHandler struct{}

func NewNoOpErrorHandler() *NoOpErrorHandler {
	return &NoOpErrorHandler{}
}

func (h *NoOpErrorHandler) Handle(ctx context.Context, err error) {
}

// The ErrorHandlerFunc type is an adapter to allow the use
// of a standard function as ErrorHandler.
type ErrorHandlerFunc func(ctx context.Context, err error)

// Handle calls f(ctx, err)
func (f ErrorHandlerFunc) Handle(ctx context.Context, err error) {
	f(ctx, err)
}
