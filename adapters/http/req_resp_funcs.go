package http

import (
	"context"
	"net/http"
)

// RequestFunc may take information from an HTTP request and put it into
// the request context. In Servers, RequestFuncs are executed before to invoke the request handler.
type RequestFunc func(context.Context, *http.Request) context.Context

// ServerResponseFunc may take information from the request context and use it
// to manipulate the ResponseWriter. ServerResponseFuncs are executed
// after invoking the request handler but before to write the response.
type ServerResponseFunc func(context.Context, http.ResponseWriter) context.Context
