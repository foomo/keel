package leaderfollower

import (
	"context"
	"encoding/json"
	"net/http"
)

// FanOutRequest is the wire format sent to each peer's /coord endpoint.
type FanOutRequest struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// FanOutResult holds the response from one peer call.
type FanOutResult struct {
	Peer   Peer
	Status int
	Body   []byte
	Err    error
}

// FanOut is provided to [Protocol.Lead] to call all current peers in parallel.
// Implementations are created by the [Coordinator] and are bound to the current
// peer discovery state.
type FanOut interface {
	// Peers returns the current peer list.
	Peers(ctx context.Context) ([]Peer, error)

	// All calls all peers concurrently with req and collects results.
	// Network-level errors are reported in FanOutResult.Err.
	// HTTP non-2xx responses are reported in FanOutResult.Status (Err is nil).
	All(ctx context.Context, req FanOutRequest) []FanOutResult
}

// Protocol defines a coordination strategy.
// The [Coordinator] calls Lead once per leadership term.
// The [Server] calls Handle for every incoming /coord request on any node.
type Protocol interface {
	// Lead is called on the leader for the duration of its leadership term.
	// fanOut allows the protocol to reach all peers without knowing about
	// discovery or transport details.
	// Lead must return when ctx is cancelled (leadership lost or shutting down).
	Lead(ctx context.Context, fanOut FanOut) error

	// Handle processes an incoming coordination message on any node.
	// action is the protocol-defined verb extracted from the request envelope.
	// body is the raw JSON payload (may be nil/empty).
	Handle(ctx context.Context, w http.ResponseWriter, action string, body []byte) error
}
