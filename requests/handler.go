package requests

import (
	"context"
	"slices"
)

// Handle provide uniform interface between input adapters and business logic operations
type Handler[Req, Resp any] interface {
	Handle(context.Context, Req) (Resp, error)
}

// HandlerFunc is an adapter to allow use ordinary function as request handlers.
// If hf is a function with proper signature, HandlerFunc(hf) is a [Handler] that calls hf
type HandlerFunc[Req, Resp any] func(context.Context, Req) (Resp, error)

// Handle calls hf(ctx, req)
func (hf HandlerFunc[Req, Resp]) Handle(ctx context.Context, req Req) (Resp, error) {
	return hf(ctx, req)
}

// Middleware is a chainable behaviour modifier of the [Handler]
type Middleware[Req, Resp any] func(Handler[Req, Resp]) Handler[Req, Resp]

// Chain is a helper function for composing middlewares.
// Requests will traverse the middleware in the order they are declared: the first middleware
// is treated as the outermost one.
func Chain[Req, Resp any](outer Middleware[Req, Resp], others ...Middleware[Req, Resp]) Middleware[Req, Resp] {
	return func(next Handler[Req, Resp]) Handler[Req, Resp] {
		for _, mdw := range slices.Backward(others) {
			next = mdw(next)
		}
		return outer(next)
	}
}
