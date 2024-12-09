package micro

import (
	"context"

	"github.com/nats-io/nats.go/micro"
)

// RequestFunc may take information from a MICRO request and put it into
// the request context. In Servers, RequestFuncs are executed before to invoke the request handler.
type RequestFunc func(context.Context, micro.Request) context.Context

// HandlerResponseFunc may take information from the request context and use it
// to manipulate the Publisher. HandlerResponseFunc are executed
// after invoking the request handler but before to write the response.
type HandlerResponseFunc func(context.Context, micro.Request) context.Context
