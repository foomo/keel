package nats_test

import (
	"testing"

	"github.com/foomo/keel"
	keelnats "github.com/foomo/keel/integration/nats"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Compile-time guards: the integration helpers accept any keel.Runtime, so both a
// long-lived *keel.Server and a short-lived *keel.Job can use the same instrumented
// Connect / NewJetStream.
var (
	_ func(keel.Runtime, string, ...nats.Option) (*nats.Conn, error)                         = keelnats.Connect
	_ func(keel.Runtime, *nats.Conn, ...jetstream.JetStreamOpt) (jetstream.JetStream, error) = keelnats.NewJetStream
	_ keel.Runtime                                                                           = (*keel.Server)(nil)
	_ keel.Runtime                                                                           = (*keel.Job)(nil)
)

func TestRuntimeAccepted(t *testing.T) {
	t.Parallel()
	// Intentionally empty: the package-level guards above assert at compile time
	// that Connect/NewJetStream accept a keel.Runtime and that both Server and Job
	// satisfy it. This test keeps the file in the test binary.
}
