package micro_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	microadapter "github.com/mcosta74/hexkit/adapters/nats/micro"
	kittesting "github.com/mcosta74/hexkit/internal/testing"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Response struct {
	Data    string `json:"data,omitempty"`
	Err     string `json:"err,omitempty"`
	ErrCode int    `json:"err_code,omitempty"`
}

func testRequest[Req any, Resp any](t *testing.T, c *nats.Conn, h *microadapter.Handler[Req, Resp]) Response {
	t.Helper()

	svc, err := micro.AddService(c, micro.Config{
		Name:       "MicroAdapterTest",
		Version:    "0.0.1",
		QueueGroup: "microadapter",
		Endpoint: &micro.EndpointConfig{
			Subject: "microadapter.test",
			Handler: h,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Stop()

	r, err := c.Request("microadapter.test", []byte("test"), 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	errCode, _ := strconv.Atoi(r.Header.Get(micro.ErrorCodeHeader))

	return Response{
		Data:    string(r.Data),
		Err:     r.Header.Get(micro.ErrorHeader),
		ErrCode: errCode,
	}
}

func TestHandler(t *testing.T) {
	s, c := kittesting.NewNATSServerAndConn(t)
	defer func() {
		s.Shutdown()
		s.WaitForShutdown()
	}()
	defer c.Close()

	t.Run("Decode Error", func(t *testing.T) {
		handler := microadapter.NewHandler(
			func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request) (struct{}, error) { return struct{}{}, errors.New("fail") },
			func(context.Context, micro.Request, struct{}) error { return nil },
		)

		resp := testRequest(t, c, handler)

		if want, got := "fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}

		if want, got := 500, resp.ErrCode; want != got {
			t.Errorf("unexpected response: want=%d, got=%d", want, got)
		}
	})

	t.Run("Port Error", func(t *testing.T) {
		handler := microadapter.NewHandler(
			func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errors.New("fail") },
			func(context.Context, micro.Request) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request, struct{}) error { return nil },
		)

		resp := testRequest(t, c, handler)

		if want, got := "fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}

		if want, got := 500, resp.ErrCode; want != got {
			t.Errorf("unexpected response: want=%d, got=%d", want, got)
		}
	})

	t.Run("Encode Error", func(t *testing.T) {
		handler := microadapter.NewHandler(
			func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request, struct{}) error { return errors.New("fail") },
		)

		resp := testRequest(t, c, handler)

		if want, got := "fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}

		if want, got := 500, resp.ErrCode; want != got {
			t.Errorf("unexpected response: want=%d, got=%d", want, got)
		}
	})

	t.Run("Happy Path", func(t *testing.T) {
		handler := microadapter.NewHandler(
			func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request) (struct{}, error) { return struct{}{}, nil },
			func(_ context.Context, r micro.Request, _ struct{}) error {
				_ = r.Respond([]byte("hello world"))
				return nil
			},
		)

		resp := testRequest(t, c, handler)

		if want, got := "", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}

		if want, got := 0, resp.ErrCode; want != got {
			t.Errorf("unexpected response: want=%d, got=%d", want, got)
		}

		if want, got := "hello world", resp.Data; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}
	})

	t.Run("Custom Error Encoder", func(t *testing.T) {
		handler := microadapter.NewHandler(
			func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, micro.Request, struct{}) error { return errors.New("fail") },
			microadapter.WithErrorEncoder[struct{}, struct{}](func(_ context.Context, err error, msg micro.Request) {
				_ = msg.Error("500", fmt.Sprintf("custom - %s", err.Error()), nil)
			}),
		)

		resp := testRequest(t, c, handler)

		if want, got := "custom - fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}

		if want, got := 500, resp.ErrCode; want != got {
			t.Errorf("unexpected response: want=%d, got=%d", want, got)
		}
	})

}
