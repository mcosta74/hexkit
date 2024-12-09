package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mcosta74/hexkit/adapters"
	"github.com/mcosta74/hexkit/ports"
)

// Server wraps a port and implement a http.Handler.
type Server[Req, Resp any] struct {
	p            ports.Port[Req, Resp]
	dec          DecodeRequestFunc[Req]
	enc          EncodeResponseFunc[Resp]
	before       []RequestFunc
	after        []ServerResponseFunc
	errorEncoder ErrorEncoder
	errorHandler adapters.ErrorHandler
}

// NewServer creates a new server, which wraps the provided port and implements http.Handler.
func NewServer[Req, Resp any](
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
type ServerOption[Req, Resp any] func(s *Server[Req, Resp])

// WithErrorEncoder sets the error encoder for the server.
func WithErrorEncoder[Req, Resp any](ee ErrorEncoder) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.errorEncoder = ee
	}
}

// WithErrorHandler sets the error handler for the server.
func WithErrorHandler[Req, Resp any](eh adapters.ErrorHandler) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.errorHandler = eh
	}
}

// WithErrorLogger sets a error handler for the server that logs errors.
func WithErrorLogger[Req, Resp any](logger *slog.Logger) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.errorHandler = adapters.NewSlogErrorHandler(logger)
	}
}

// WithServerBefore functions are executed on the HTTP request object
// before the port is invoked.
func WithServerBefore[Req, Resp any](before ...RequestFunc) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.before = append(s.before, before...)
	}
}

// WithServerAfter functions are executed on the HTTP response writer
// after the port is invoked, but before anything is written on the client.
func WithServerAfter[Req, Resp any](after ...ServerResponseFunc) ServerOption[Req, Resp] {
	return func(s *Server[Req, Resp]) {
		s.after = append(s.after, after...)
	}
}

// ServeHTTP implements http.Handler.
func (s Server[Req, Resp]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	for _, f := range s.before {
		ctx = f(ctx, r)
	}

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

	for _, f := range s.after {
		ctx = f(ctx, w)
	}

	if err := s.enc(ctx, w, response); err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, w)
		return
	}
}

// ErrorEncoder encodes an error to the ResponseWriter.
type ErrorEncoder func(ctx context.Context, err error, w http.ResponseWriter)

// DefaultErrorEncoder is used when no error encoder is provided.
func DefaultErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plan")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

// JSONErrorEncoder encodes errors in JSON format.
func JSONErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	data := struct {
		Err string `json:"err,omitempty"`
	}{
		Err: err.Error(),
	}

	body, _ := json.Marshal(data)
	_, _ = w.Write(body)
}

// NoOpRequestDecoder it's a decoder that does nothing
func NoOpRequestDecoder[Req any](context.Context, *http.Request) (Req, error) {
	var req Req
	return req, nil
}
