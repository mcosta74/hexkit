package shared

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func NewNATSServerAndConn(t *testing.T) (*server.Server, *nats.Conn) {
	s, err := server.NewServer(&server.Options{
		Host: "localhost",
		Port: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	go s.Start()

	const tries = 10
	for i := 0; i < tries && !s.Running(); i++ {
		t.Logf("Running %v", s.Running())
		time.Sleep(time.Second)
	}
	if !s.Running() {
		s.Shutdown()
		s.WaitForShutdown()
		t.Fatal("not yet running")
	}

	if !s.ReadyForConnections(tries * time.Second) {
		log.Fatal("NATS server not ready in time")
	}

	c, err := nats.Connect(fmt.Sprintf("nats://%s", s.Addr().String()), nats.Name(t.Name()))
	if err != nil {
		t.Fatalf("error connecting to the server: %s", err)
	}
	return s, c
}
