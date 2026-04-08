package nats_test

import (
	"fmt"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Event struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// startNATSServer starts an embedded NATS server on a random port and returns
// a connected client. Panics on failure (suitable for example tests).
func startNATSServer() (*natsserver.Server, *nats.Conn) {
	opts := &natsserver.Options{
		Host:  "127.0.0.1",
		Port:  -1,
		NoLog: true,
	}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		panic(fmt.Sprintf("nats server: %s", err))
	}
	srv.Start()
	if !srv.ReadyForConnections(5e9) { // 5s
		panic("nats server not ready")
	}
	conn, err := nats.Connect(srv.ClientURL())
	if err != nil {
		srv.Shutdown()
		panic(fmt.Sprintf("nats connect: %s", err))
	}
	return srv, conn
}
