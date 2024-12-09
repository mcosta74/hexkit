package nats_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	natsadapter "github.com/mcosta74/hexkit/adapters/nats"
	kittesting "github.com/mcosta74/hexkit/internal/testing"
	"github.com/mcosta74/hexkit/requests"
)

type Response struct {
	Data string `json:"data,omitempty"`
	Err  string `json:"err,omitempty"`
}

func testRequest[Req, Resp any](t *testing.T, c *nats.Conn, h *natsadapter.Subscriber[Req, Resp]) Response {
	t.Helper()

	sub, err := c.QueueSubscribe("natsadapter.test", "natsadapter", h.ServeMsg(c))
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()

	r, err := c.Request("natsadapter.test", []byte("test"), 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	var resp Response
	err = json.Unmarshal(r.Data, &resp)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestSubscriber(t *testing.T) {
	s, c := kittesting.NewNATSServerAndConn(t)
	defer func() {
		s.Shutdown()
		s.WaitForShutdown()
	}()
	defer c.Close()

	t.Run("Decode Error", func(t *testing.T) {
		handler := natsadapter.NewSubscriber(
			requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil }),
			func(context.Context, *nats.Msg) (struct{}, error) { return struct{}{}, errors.New("fail") },
			func(context.Context, string, *nats.Conn, struct{}) error { return nil },
		)

		resp := testRequest(t, c, handler)

		if want, got := "fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}
	})
	t.Run("Handler Error", func(t *testing.T) {
		handler := natsadapter.NewSubscriber(
			requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errors.New("fail") }),
			func(context.Context, *nats.Msg) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, string, *nats.Conn, struct{}) error { return nil },
		)

		resp := testRequest(t, c, handler)

		if want, got := "fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}

	})
	t.Run("Encode Error", func(t *testing.T) {
		handler := natsadapter.NewSubscriber(
			requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil }),
			func(context.Context, *nats.Msg) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, string, *nats.Conn, struct{}) error { return errors.New("fail") },
		)

		resp := testRequest(t, c, handler)

		if want, got := "fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}
	})

	t.Run("Happy Path", func(t *testing.T) {
		handler := natsadapter.NewSubscriber(
			requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, nil }),
			func(context.Context, *nats.Msg) (struct{}, error) { return struct{}{}, nil },
			func(_ context.Context, reply string, nc *nats.Conn, _ struct{}) error {
				response := struct {
					Data string `json:"data,omitempty"`
				}{
					Data: "hello world",
				}

				b, err := json.Marshal(response)
				if err != nil {
					return err
				}
				_ = nc.Publish(reply, b)
				return nil
			},
		)

		resp := testRequest(t, c, handler)

		if want, got := "hello world", resp.Data; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}
	})

	t.Run("Custom Error Encoder", func(t *testing.T) {
		handler := natsadapter.NewSubscriber(
			requests.HandlerFunc[struct{}, struct{}](func(context.Context, struct{}) (struct{}, error) { return struct{}{}, errors.New("fail") }),
			func(context.Context, *nats.Msg) (struct{}, error) { return struct{}{}, nil },
			func(context.Context, string, *nats.Conn, struct{}) error { return nil },
			natsadapter.WithErrorEncoder[struct{}, struct{}](func(ctx context.Context, err error, reply string, nc *nats.Conn) {
				response := struct {
					Error string `json:"err,omitempty"`
				}{
					Error: fmt.Sprintf("custom - %s", err.Error()),
				}

				b, err := json.Marshal(response)
				if err != nil {
					return
				}
				_ = nc.Publish(reply, b)
			}),
		)

		resp := testRequest(t, c, handler)

		if want, got := "custom - fail", resp.Err; want != got {
			t.Errorf("unexpected response: want=%q, got=%q", want, got)
		}
	})

}
