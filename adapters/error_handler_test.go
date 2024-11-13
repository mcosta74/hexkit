package adapters

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestSlogErrorHandler(t *testing.T) {
	writer := &bytes.Buffer{}

	logger := slog.New(slog.NewTextHandler(
		writer,
		&slog.HandlerOptions{
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		},
	))

	errorHandler := NewSlogErrorHandler(logger)
	err := errors.New("failure")

	errorHandler.Handle(context.Background(), err)

	want := "level=ERROR msg=error err=failure\n"
	if got := writer.String(); got != want {
		t.Errorf("expected log record got %q, want %q", got, want)
	}
}
