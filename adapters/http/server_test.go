package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	kithttp "github.com/mcosta74/hexkit/adapters/http"
	"github.com/mcosta74/hexkit/requests"
)

func checkResponse(t *testing.T, resp *http.Response, wantCode int, wantBody []byte) {
	t.Helper()

	if want, got := wantCode, resp.StatusCode; want != got {
		t.Errorf("want: %d, got: %d", want, got)
	}
	defer resp.Body.Close()

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error reading response body: %v", err)
	}

	if string(wantBody) != string(got) {
		t.Errorf("unexpected body, want: %q, got: %q", string(wantBody), string(got))
	}

}

func TestServerDecodeError(t *testing.T) {
	handler := kithttp.NewServer(
		requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil }),
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
	checkResponse(t, resp, http.StatusInternalServerError, []byte("fail"))
}

func TestServerPortError(t *testing.T) {
	handler := kithttp.NewServer(
		requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errors.New("fail") }),
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
	checkResponse(t, resp, http.StatusInternalServerError, []byte("fail"))
}

func TestServerEncodeError(t *testing.T) {
	handler := kithttp.NewServer(
		requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil }),
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, nil },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			return errors.New("fail")
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)
	checkResponse(t, resp, http.StatusInternalServerError, []byte("fail"))
}

func TestServerNoError(t *testing.T) {
	handler := kithttp.NewServer(
		requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil }),
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
	checkResponse(t, resp, http.StatusOK, []byte("ok"))
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
		requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errValidation }),
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
	checkResponse(t, resp, http.StatusBadRequest, []byte(""))
}

func TestServerJSONErrorEncoder(t *testing.T) {
	handler := kithttp.NewServer(
		requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errors.New("fail") }),
		func(ctx context.Context, r *http.Request) (struct{}, error) { return struct{}{}, nil },
		func(_ context.Context, w http.ResponseWriter, _ struct{}) error {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return nil
		},
		kithttp.WithErrorEncoder[struct{}, struct{}](kithttp.JSONErrorEncoder),
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, _ := http.Get(server.URL)

	expectedBody := struct {
		Err string `json:"err,omitempty"`
	}{
		Err: "fail",
	}

	d, _ := json.Marshal(expectedBody)
	checkResponse(t, resp, http.StatusInternalServerError, d)
}
