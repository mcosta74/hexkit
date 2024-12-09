package ports

import (
	"context"
	"slices"
)

// Port is the fundamental building block of servers and clients.
// It represents a single RPC method.
type Port[Req, Resp any] func(ctx context.Context, request Req) (response Resp, err error)

// Middleware is a chainable behavior modifier for the ports.
type Middleware[Req, Resp any] func(next Port[Req, Resp]) Port[Req, Resp]

// Chain is a helper function for composing middlewares.
// Requests will traverse the middleware in the order they are declared: the first middleware
// is treated as the outermost one.
func Chain[Req, Resp any](outer Middleware[Req, Resp], others ...Middleware[Req, Resp]) Middleware[Req, Resp] {
	return func(next Port[Req, Resp]) Port[Req, Resp] {
		for _, mdw := range slices.Backward(others) {
			next = mdw(next)
		}
		return outer(next)
	}
}
