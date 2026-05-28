package threephase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	lf "github.com/foomo/keel/service/leaderfollower"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultLeaseName     = "leaderfollower"
	defaultConfigMapName = "leaderfollower"
	defaultCoordPort     = 8090
	defaultCoordPath     = "/coord"
)

// K8sConfig holds Kubernetes infrastructure configuration for [New].
// Namespace and PodName are auto-detected if left empty (from the K8s service
// account token and hostname, respectively). LabelSel is required.
type K8sConfig[S any] struct {
	// Required
	LabelSel string // e.g. "app.kubernetes.io/instance=myservice"

	// Optional — auto-detected if empty
	Namespace string // read from /var/run/secrets/kubernetes.io/serviceaccount/namespace
	PodName   string // read from os.Hostname()

	// Optional — defaults shown in parentheses
	LeaseName        string       // ("leaderfollower") Kubernetes Lease resource name
	ConfigMapName    string       // ("leaderfollower") ConfigMap for state persistence
	CoordPort        int          // (8090) port the coord HTTP server listens on
	CoordPath        string       // ("/coord") path of the coord endpoint
	CoordServiceName string       // ("") K8s Service name for ProposeURL(); defaults to LeaseName
	HTTPClient       *http.Client // custom HTTP client for fan-out calls; default uses 5s timeout
}

func (cfg *K8sConfig[S]) applyDefaults() error {
	if cfg.Namespace == "" {
		if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			cfg.Namespace = strings.TrimSpace(string(data))
		}
	}

	if cfg.Namespace == "" {
		return fmt.Errorf("namespace not set and could not be auto-detected from service account token")
	}

	if cfg.PodName == "" {
		cfg.PodName, _ = os.Hostname()
	}

	if cfg.LeaseName == "" {
		cfg.LeaseName = defaultLeaseName
	}

	if cfg.ConfigMapName == "" {
		cfg.ConfigMapName = defaultConfigMapName
	}

	if cfg.CoordPort == 0 {
		cfg.CoordPort = defaultCoordPort
	}

	if cfg.CoordPath == "" {
		cfg.CoordPath = defaultCoordPath
	}

	if cfg.CoordServiceName == "" {
		cfg.CoordServiceName = cfg.LeaseName
	}

	return nil
}

// CoordinatorHandle is the batteries-included entry point returned by [New].
// It wires a [lf.Coordinator] (election + fan-out) together with a coord HTTP
// server, and exposes generic state access via [CommittedState] and [ProposeURL].
type CoordinatorHandle[S any] struct {
	l          *zap.Logger
	coord      *lf.Coordinator
	server     *lf.Server
	store      CoordStore[S]
	proposeURL string
}

// Name implements the keel Service interface.
func (h *CoordinatorHandle[S]) Name() string { return h.coord.Name() }

// IsLeader reports whether this node currently holds the leader lease.
func (h *CoordinatorHandle[S]) IsLeader() bool { return h.coord.IsLeader() }

// Start begins the coord HTTP server (in a goroutine) and the election loop
// (blocking). Cancel ctx to stop both. Call as go handle.Start(ctx) or register
// with keel.Server.
func (h *CoordinatorHandle[S]) Start(ctx context.Context) error {
	srvDone := make(chan error, 1)

	go func() {
		srvDone <- h.server.Start(ctx)
	}()

	err := h.coord.Start(ctx)

	// Election loop ended; shut down the coord HTTP server.
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = h.server.Close(shutCtx) //nolint:contextcheck // shutCtx uses context.Background because ctx is already cancelled at this point

	<-srvDone // wait for the server goroutine to exit

	return err
}

// Close shuts down the coord HTTP server gracefully.
func (h *CoordinatorHandle[S]) Close(ctx context.Context) error {
	return h.server.Close(ctx)
}

// CommittedState returns the currently committed state by reading the store.
// Returns the zero value of S when no committed state exists yet.
func (h *CoordinatorHandle[S]) CommittedState(ctx context.Context) (S, error) {
	committed, _, _, err := h.store.LoadCommitted(ctx)
	return committed, err
}

// ProposeURL returns the coord propose endpoint URL for external callers (e.g. batch jobs).
// The URL points to the K8s ClusterIP Service, making it stable across pod restarts.
// Format: http://<CoordServiceName>.<namespace>.svc.cluster.local:<port><path>
func (h *CoordinatorHandle[S]) ProposeURL() string {
	return h.proposeURL
}

// New creates a batteries-included [CoordinatorHandle] wired for Kubernetes.
//
// It builds a [lf.LeaseElector], [lf.PodDiscovery], [ConfigMapCoordStore], and a
// [Protocol], then wraps them in a CoordinatorHandle that starts both the election
// loop and the coord HTTP server when [CoordinatorHandle.Start] is called.
//
// canCommit and doCommit are required. Pass additional protocol options via opts
// (e.g. [WithPreCommit], [WithRollback], [WithLeaderAfterCommit]).
// Framework-owned concerns (ConfigMap persistence, crash recovery, proposed-state
// watching) are wired automatically.
func New[S any](
	l *zap.Logger,
	client kubernetes.Interface,
	cfg K8sConfig[S],
	canCommit func(ctx context.Context, proposed S) error,
	doCommit func(ctx context.Context, proposed S) error,
	opts ...Option[S],
) (*CoordinatorHandle[S], error) {
	if err := cfg.applyDefaults(); err != nil {
		return nil, err
	}

	elector := lf.NewLeaseElector(lf.LeaseElectorConfig{
		Client:    client,
		Namespace: cfg.Namespace,
		LeaseName: cfg.LeaseName,
		Identity:  cfg.PodName,
	})

	discovery := lf.NewPodDiscovery(lf.PodDiscoveryConfig{
		Client:        client,
		Namespace:     cfg.Namespace,
		LabelSelector: cfg.LabelSel,
		CoordPort:     cfg.CoordPort,
	})

	store := NewConfigMapCoordStore[S](client, cfg.Namespace, cfg.ConfigMapName, l)

	// Create the protocol with user opts first so we can read back any
	// WithLeaderAfterCommit hook the caller set before wrapping it with the
	// framework's ConfigMap persistence.
	proto := NewProtocol(l, store, canCommit, doCommit, opts...)
	userAfterCommit := proto.afterCommit // may be nil

	// Framework afterCommit: always persist committed/previous state first, then
	// call the user's WithLeaderAfterCommit hook if one was provided.
	proto.afterCommit = func(ctx context.Context, previous, committed S) error {
		if err := store.SetCommitted(ctx, committed, previous); err != nil {
			l.Error("threephase: failed to persist committed state to ConfigMap", zap.Error(err))
		}

		if userAfterCommit != nil {
			return userAfterCommit(ctx, previous, committed)
		}

		return nil
	}

	addr := fmt.Sprintf(":%d", cfg.CoordPort)

	var coordOpts []lf.CoordinatorOption
	if cfg.HTTPClient != nil {
		coordOpts = append(coordOpts, lf.WithHTTPClient(cfg.HTTPClient))
	}

	coordOpts = append(coordOpts, lf.WithCoordPath(cfg.CoordPath))

	coordinator := lf.NewCoordinator(l, elector, discovery, proto, coordOpts...)

	server := lf.NewServer(l, proto, addr,
		lf.WithServerCoordPath(cfg.CoordPath),
	)

	proposeURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d%s",
		cfg.CoordServiceName, cfg.Namespace, cfg.CoordPort, cfg.CoordPath)

	return &CoordinatorHandle[S]{
		l:          l,
		coord:      coordinator,
		server:     server,
		store:      store,
		proposeURL: proposeURL,
	}, nil
}

// defaultProposeClient is used by [Propose] when no custom client is provided.
// The 10-second timeout prevents hangs when the coord server is unavailable;
// a context deadline passed by the caller will take precedence when shorter.
var defaultProposeClient = &http.Client{Timeout: 10 * time.Second}

// Propose sends a propose action to the coord server at coordURL, triggering
// a new commit round for the given state. Use this in external services (e.g.
// batch jobs) instead of patching the ConfigMap directly.
//
// coordURL is typically obtained from [CoordinatorHandle.ProposeURL] and passed
// as an environment variable to the external caller.
func Propose[S any](ctx context.Context, coordURL string, state S) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal propose payload: %w", err)
	}

	body, err := json.Marshal(lf.FanOutRequest{
		Action:  actionPropose,
		Payload: json.RawMessage(payload),
	})
	if err != nil {
		return fmt.Errorf("marshal propose envelope: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, coordURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build propose request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := defaultProposeClient.Do(req)
	if err != nil {
		return fmt.Errorf("propose request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("propose: unexpected status %d", resp.StatusCode)
	}

	return nil
}
