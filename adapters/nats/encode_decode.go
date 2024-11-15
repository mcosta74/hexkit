package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

// DecodeRequestFunc extracts user-domain request object from a publisher request object.
type DecodeRequestFunc[Req any] func(ctx context.Context, msg *nats.Msg) (request Req, err error)

// EncodeResponseFunc encodes the provided response object to the subscriber reply.
type EncodeResponseFunc[Resp any] func(ctx context.Context, subject string, nc *nats.Conn, resp Resp) error
