package testing

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func NewNATSServerAndConn(t *testing.T) (*server.Server, *nats.Conn) {
	t.Helper()

	s, err := server.NewServer(&server.Options{
		Host: "localhost",
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		t.Fatal(err)
	}

	go s.Start()

	if !s.ReadyForConnections(5 * time.Second) {
		log.Fatal("NATS server not ready in time")
	}

	c, err := nats.Connect(fmt.Sprintf("nats://%s", s.Addr().String()), nats.Name(t.Name()))
	if err != nil {
		t.Fatalf("error connecting to the server: %s", err)
	}
	return s, c
}
