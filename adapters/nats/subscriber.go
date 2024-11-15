package nats

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mcosta74/hexkit/adapters"
	"github.com/mcosta74/hexkit/ports"
	"github.com/nats-io/nats.go"
)

// Subscriber wraps a port and provides nats.MsgHandler.
type Subscriber[Req any, Resp any] struct {
	p            ports.Port[Req, Resp]
	dec          DecodeRequestFunc[Req]
	enc          EncodeResponseFunc[Resp]
	errorEncoder ErrorEncoder
	errorHandler adapters.ErrorHandler
}

// NewServer creates a new subscriber, which wraps the provided port and provides a nats.MsgHandler.
func NewSubscriber[Req any, Resp any](
	p ports.Port[Req, Resp],
	dec DecodeRequestFunc[Req],
	enc EncodeResponseFunc[Resp],
	options ...SubscriberOption[Req, Resp],

) *Subscriber[Req, Resp] {
	s := &Subscriber[Req, Resp]{
		p:            p,
		dec:          dec,
		enc:          enc,
		errorEncoder: DefaultErrorEncoder,
		errorHandler: adapters.NewNoOpErrorHandler(),
	}

	for _, o := range options {
		o(s)
	}
	return s
}

// SubscriberOption sets optional parameter for the subscriber.
type SubscriberOption[Req any, Resp any] func(s *Subscriber[Req, Resp])

// WithErrorEncoder sets the error encoder for the subscriber.
func WithErrorEncoder[Req any, Resp any](ee ErrorEncoder) SubscriberOption[Req, Resp] {
	return func(s *Subscriber[Req, Resp]) {
		s.errorEncoder = ee
	}
}

// WithErrorHandler sets the error handler for the subscriber.
func WithErrorHandler[Req any, Resp any](eh adapters.ErrorHandler) SubscriberOption[Req, Resp] {
	return func(s *Subscriber[Req, Resp]) {
		s.errorHandler = eh
	}
}

// WithErrorLogger sets a error handler for the subscriber that logs errors.
func WithErrorLogger[Req any, Resp any](logger *slog.Logger) SubscriberOption[Req, Resp] {
	return func(s *Subscriber[Req, Resp]) {
		s.errorHandler = adapters.NewSlogErrorHandler(logger)
	}
}

// ServeMsg provides nats.MsgHandler
func (s *Subscriber[Req, Resp]) ServeMsg(nc *nats.Conn) nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		request, err := s.dec(ctx, msg)
		if err != nil {
			s.errorHandler.Handle(ctx, err)
			if msg.Reply != "" {
				s.errorEncoder(ctx, err, msg.Reply, nc)
			}
			return
		}

		response, err := s.p(ctx, request)
		if err != nil {
			s.errorHandler.Handle(ctx, err)
			if msg.Reply != "" {
				s.errorEncoder(ctx, err, msg.Reply, nc)
			}
			return
		}

		if msg.Reply != "" {
			if err := s.enc(ctx, msg.Reply, nc, response); err != nil {
				s.errorHandler.Handle(ctx, err)
				s.errorEncoder(ctx, err, msg.Reply, nc)
				return
			}
		}
	}
}

// ErrorEncoder encodes an error to the subscriber reply.
type ErrorEncoder func(ctx context.Context, err error, reply string, nc *nats.Conn)

// DefaultErrorEncoder is used when no error encoder is provided
func DefaultErrorEncoder(ctx context.Context, err error, reply string, nc *nats.Conn) {
	response := struct {
		Error string `json:"err,omitempty"`
	}{
		Error: err.Error(),
	}

	b, err := json.Marshal(response)
	if err != nil {
		return
	}
	_ = nc.Publish(reply, b)
}
