package micro

import (
	"context"
	"log/slog"

	"github.com/mcosta74/hexkit/adapters"
	"github.com/mcosta74/hexkit/ports"
	"github.com/nats-io/nats.go/micro"
)

// Handler wraps a port and implements micro.Handler.
type Handler[Req any, Resp any] struct {
	p            ports.Port[Req, Resp]
	dec          DecodeRequestFunc[Req]
	enc          EncodeResponseFunc[Resp]
	errorEncoder ErrorEncoder
	errorHandler adapters.ErrorHandler
}

// NewHandler creates a new handler, which wraps the provided port and implements a micro.Handler.
func NewHandler[Req any, Resp any](
	p ports.Port[Req, Resp],
	dec DecodeRequestFunc[Req],
	enc EncodeResponseFunc[Resp],
	options ...HandlerOption[Req, Resp],

) *Handler[Req, Resp] {
	s := &Handler[Req, Resp]{
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

// HandlerOption sets optional parameter for the handler.
type HandlerOption[Req any, Resp any] func(s *Handler[Req, Resp])

// WithErrorEncoder sets the error encoder for the handler.
func WithErrorEncoder[Req any, Resp any](ee ErrorEncoder) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.errorEncoder = ee
	}
}

// WithErrorHandler sets the error handler for the handler.
func WithErrorHandler[Req any, Resp any](eh adapters.ErrorHandler) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.errorHandler = eh
	}
}

// WithErrorLogger sets a error handler for the handler that logs errors.
func WithErrorLogger[Req any, Resp any](logger *slog.Logger) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.errorHandler = adapters.NewSlogErrorHandler(logger)
	}
}

// Handle implements micro.Handler
func (s *Handler[Req, Resp]) Handle(msg micro.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request, err := s.dec(ctx, msg)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		if msg.Reply() != "" {
			s.errorEncoder(ctx, err, msg)
		}
		return
	}

	response, err := s.p(ctx, request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		if msg.Reply() != "" {
			s.errorEncoder(ctx, err, msg)
		}
		return
	}

	if msg.Reply() != "" {
		if err := s.enc(ctx, msg, response); err != nil {
			s.errorHandler.Handle(ctx, err)
			s.errorEncoder(ctx, err, msg)
			return
		}
	}
}

// ErrorEncoder encodes an error to the handler reply.
type ErrorEncoder func(ctx context.Context, err error, msg micro.Request)

// DefaultErrorEncoder is used when no error encoder is provided
func DefaultErrorEncoder(ctx context.Context, err error, msg micro.Request) {
	_ = msg.Respond([]byte(err.Error()))
}
