package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/mcosta74/hexkit/adapters"
	"github.com/mcosta74/hexkit/ports"
)

// Server wraps a port and implement a http.Handler.
type Server[Req any, Resp any] struct {
	p            ports.Port[Req, Resp]
	dec          DecodeRequestFunc[Req]
	enc          EncodeResponseFunc[Resp]
	errorEncoder ErrorEncoder
	errorHandler adapters.ErrorHandler
}

// NewServer creates a new server, which wraps the provided port and implements http.Handler.
func NewServer[Req any, Resp any](
	p ports.Port[Req, Resp],
	dec DecodeRequestFunc[Req],
	enc EncodeResponseFunc[Resp],
	options ...ServerOption[Req, Resp],
) *Server[Req, Resp] {
	s := &Server[Req, Resp]{
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

// ServerOption sets optional parameter for the server.
type ServerOption[Req any, Resp any] func(s *Server[Req, Resp])

// WithErrorEncoder sets the error encoder for the server.
func WithErrorEncoder[Req any, Resp any](ee ErrorEncoder) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.errorEncoder = ee
	}
}

// WithErrorHandler sets the error handler for the server.
func WithErrorHandler[Req any, Resp any](eh adapters.ErrorHandler) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.errorHandler = eh
	}
}

// WithErrorLogger sets a error handler for the server that logs errors.
func WithErrorLogger[Req any, Resp any](logger *slog.Logger) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.errorHandler = adapters.NewSlogErrorHandler(logger)
	}
}

// ServeHTTP implements http.Handler.
func (s Server[Req, Resp]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := s.dec(ctx, r)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, w)
		return
	}

	response, err := s.p(ctx, request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, w)
		return
	}

	if err := s.enc(ctx, w, response); err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, w)
		return
	}
}

// ErrorEncoder encodes an error to the ResponseWriter.
type ErrorEncoder func(ctx context.Context, err error, w http.ResponseWriter)

// DefaultErrorEncoder is used when no error encoder is provided
func DefaultErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plan")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}
