package http

import (
	"context"
	"net/http"
)

// DecodeRequestFunc extracts user-domain request object from an HTTP request object.
type DecodeRequestFunc[Req any] func(ctx context.Context, r *http.Request) (request Req, err error)

// EncodeResponseFunc encodes the provided response object to the HTTP response writer.
type EncodeResponseFunc[Resp any] func(ctx context.Context, rw http.ResponseWriter, resp Resp) error
