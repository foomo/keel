package leaderfollower

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const defaultCoordPath = "/coord"
const defaultHTTPTimeout = 5 * time.Second

// Coordinator wires together an [Elector], [PeerDiscovery], and [Protocol].
type Coordinator struct {
	l         *zap.Logger
	elector   Elector
	discovery PeerDiscovery
	protocol  Protocol
	isLeader  atomic.Bool

	httpClient *http.Client
	coordPath  string
}

// CoordinatorOption configures a [Coordinator].
type CoordinatorOption func(*Coordinator)

// WithHTTPClient replaces the default HTTP client used for fan-out calls.
func WithHTTPClient(c *http.Client) CoordinatorOption {
	return func(coord *Coordinator) { coord.httpClient = c }
}

// WithCoordPath sets the path used for peer coordination requests.
// Default: "/coord".
func WithCoordPath(path string) CoordinatorOption {
	return func(coord *Coordinator) { coord.coordPath = path }
}

// NewCoordinator creates a Coordinator from explicitly provided components.
func NewCoordinator(l *zap.Logger, elector Elector, discovery PeerDiscovery, protocol Protocol, opts ...CoordinatorOption) *Coordinator {
	c := &Coordinator{
		l:         l,
		elector:   elector,
		discovery: discovery,
		protocol:  protocol,
		coordPath: defaultCoordPath,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
	for _, o := range opts {
		o(c)
	}

	return c
}

// Name implements the keel Service interface.
func (c *Coordinator) Name() string { return "leaderfollower-coordinator" }

// Start runs the election loop until ctx is cancelled.
// Blocks
func (c *Coordinator) Start(ctx context.Context) error {
	return c.elector.Run(ctx, func(leaderCtx context.Context) {
		c.isLeader.Store(true)
		c.l.Info("became leader", zap.String("identity", c.elector.Identity()))

		defer func() {
			c.isLeader.Store(false)
			c.l.Info("lost leadership", zap.String("identity", c.elector.Identity()))
		}()

		if err := c.protocol.Lead(leaderCtx, c.newFanOut()); err != nil {
			c.l.Error("protocol lead returned error", zap.Error(err))
		}
	})
}

// Close is a no-op; cancelling the context passed to Start is sufficient.
func (c *Coordinator) Close(_ context.Context) error { return nil }

// IsLeader reports whether this node currently holds the leader lease.
func (c *Coordinator) IsLeader() bool { return c.isLeader.Load() }

// newFanOut returns a FanOut implementation bound to this coordinator's
// peer discovery and HTTP client.
func (c *Coordinator) newFanOut() FanOut {
	return &fanOutImpl{
		l:         c.l,
		discovery: c.discovery,
		client:    c.httpClient,
		path:      c.coordPath,
	}
}

// fanOutImpl is the concrete FanOut used by protocols.
type fanOutImpl struct {
	l         *zap.Logger
	discovery PeerDiscovery
	client    *http.Client
	path      string
}

func (f *fanOutImpl) Peers(ctx context.Context) ([]Peer, error) {
	return f.discovery.Peers(ctx)
}

func (f *fanOutImpl) All(ctx context.Context, req FanOutRequest) []FanOutResult {
	peers, err := f.discovery.Peers(ctx)
	if err != nil {
		f.l.Error("fan-out: peer discovery failed", zap.Error(err))
		return []FanOutResult{{Err: fmt.Errorf("%w: %w", ErrDiscoveryFailed, err)}}
	}

	body, err := json.Marshal(req)
	if err != nil {
		f.l.Error("fan-out: failed to marshal request", zap.Error(err))
		return []FanOutResult{{Err: fmt.Errorf("marshal fan-out request: %w", err)}}
	}

	var mu sync.Mutex

	results := make([]FanOutResult, 0, len(peers))

	var wg sync.WaitGroup

	for _, p := range peers {
		wg.Go(func() {
			res := f.callPeer(ctx, p, body)

			mu.Lock()

			results = append(results, res)
			mu.Unlock()
		})
	}

	wg.Wait()

	return results
}

func (f *fanOutImpl) callPeer(ctx context.Context, p Peer, body []byte) FanOutResult {
	url := p.Addr + f.path

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return FanOutResult{Peer: p, Err: err}
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return FanOutResult{Peer: p, Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return FanOutResult{Peer: p, Status: resp.StatusCode, Err: fmt.Errorf("read response: %w", err)}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return FanOutResult{Peer: p, Status: resp.StatusCode,
			Err: fmt.Errorf("%w: HTTP %d from %s", ErrPeerCallFailed, resp.StatusCode, p.ID), Body: respBody}
	}

	return FanOutResult{Peer: p, Status: resp.StatusCode, Body: respBody}
}
