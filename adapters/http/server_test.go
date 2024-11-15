package http_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	kithttp "github.com/mcosta74/hexkit/adapters/http"
)

func TestServerDecodeError(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, errors.New("fail") },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return nil
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)
	if want, got := http.StatusInternalServerError, resp.StatusCode; want != got {
		t.Errorf("want: %d, got: %d", want, got)
	}
}

func TestServerPortError(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errors.New("fail") },
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, nil },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return nil
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)
	if want, got := http.StatusInternalServerError, resp.StatusCode; want != got {
		t.Errorf("want: %d, got: %d", want, got)
	}
}

func TestServerEncodeError(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, nil },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			return errors.New("fail")
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)
	if want, got := http.StatusInternalServerError, resp.StatusCode; want != got {
		t.Errorf("want: %d, got: %d", want, got)
	}
}

func TestServerNoError(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, nil },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return nil
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)
	if want, got := http.StatusOK, resp.StatusCode; want != got {
		t.Errorf("not expected status_code: want: %d, got: %d", want, got)
	}
	defer resp.Body.Close()

	buf, _ := io.ReadAll(resp.Body)
	if want, got := "ok", string(buf); want != got {
		t.Errorf("not expected body: want: %q, got: %q", want, got)

	}
}

func TestServerErrorEncoder(t *testing.T) {
	errValidation := errors.New("validation")
	errCode := func(err error) int {
		if errors.Is(err, errValidation) {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}

	handler := kithttp.NewServer(
		func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errValidation },
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, nil },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return nil
		},
		kithttp.WithErrorEncoder[struct{}, struct{}](func(ctx context.Context, err error, w http.ResponseWriter) {
			w.WriteHeader(errCode(err))
		}),
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)
	if want, got := http.StatusBadRequest, resp.StatusCode; want != got {
		t.Errorf("not expected status_code: want: %d, got: %d", want, got)
	}
}
