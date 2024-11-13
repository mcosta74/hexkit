package micro

import (
	"context"

	"github.com/nats-io/nats.go/micro"
)

// DecodeRequestFunc extracts user-domain request object from a publisher request object.
type DecodeRequestFunc[Req any] func(ctx context.Context, msg micro.Request) (request Req, err error)

// EncodeResponseFunc encodes the provided response object to the subscriber reply.
type EncodeResponseFunc[Resp any] func(ctx context.Context, msg micro.Request, resp Resp) error
