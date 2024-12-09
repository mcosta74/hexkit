package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

// RequestFunc may take information from a NATS message and put it into
// the request context. In Servers, RequestFuncs are executed before to invoke the request handler.
type RequestFunc func(context.Context, *nats.Msg) context.Context

// SubscriberResponseFunc may take information from the request context and use it
// to manipulate the Publisher. SubscriberResponseFuncs are executed
// after invoking the request handler but before to write the response.
type SubscriberResponseFunc func(context.Context, *nats.Conn) context.Context
