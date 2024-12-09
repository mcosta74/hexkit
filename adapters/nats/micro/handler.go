package micro

import (
	"context"
	"log/slog"

	"github.com/mcosta74/hexkit/adapters"
	"github.com/mcosta74/hexkit/ports"
	"github.com/nats-io/nats.go/micro"
)

// Handler wraps a port and implements micro.Handler.
type Handler[Req, Resp any] struct {
	p            ports.Port[Req, Resp]
	dec          DecodeRequestFunc[Req]
	enc          EncodeResponseFunc[Resp]
	before       []RequestFunc
	after        []HandlerResponseFunc
	errorEncoder ErrorEncoder
	errorHandler adapters.ErrorHandler
}

// NewHandler creates a new handler, which wraps the provided port and implements a micro.Handler.
func NewHandler[Req, Resp any](
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
type HandlerOption[Req, Resp any] func(s *Handler[Req, Resp])

// WithErrorEncoder sets the error encoder for the handler.
func WithErrorEncoder[Req, Resp any](ee ErrorEncoder) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.errorEncoder = ee
	}
}

// WithErrorHandler sets the error handler for the handler.
func WithErrorHandler[Req, Resp any](eh adapters.ErrorHandler) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.errorHandler = eh
	}
}

// WithErrorLogger sets a error handler for the handler that logs errors.
func WithErrorLogger[Req, Resp any](logger *slog.Logger) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.errorHandler = adapters.NewSlogErrorHandler(logger)
	}
}

// WithHandlerBefore functions are executed on the NATS message object
// before the port is invoked.
func WithHandlerBefore[Req, Resp any](before ...RequestFunc) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.before = append(s.before, before...)
	}
}

// WithHandlerAfter functions are executed on the HTTP response writer
// after the port is invoked, but before anything is written on the client.
func WithHandlerAfter[Req, Resp any](after ...HandlerResponseFunc) HandlerOption[Req, Resp] {
	return func(s *Handler[Req, Resp]) {
		s.after = append(s.after, after...)
	}
}

// Handle implements micro.Handler
func (s *Handler[Req, Resp]) Handle(msg micro.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, f := range s.before {
		ctx = f(ctx, msg)
	}

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

	for _, f := range s.after {
		ctx = f(ctx, msg)
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
	_ = msg.Error("500", err.Error(), nil)
}

// NoOpRequestDecoder it's a decoder that does nothing
func NoOpRequestDecoder[Req any](context.Context, micro.Request) (Req, error) {
	var req Req
	return req, nil
}
