package leaderfollower

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// Elector manages distributed leader election.
// The winner calls onLeading with a context that is cancelled when leadership
// is lost. Implementations must return from Run when ctx is cancelled.
type Elector interface {
	// Identity returns the stable node identifier used for election.
	// It must match [Peer.ID] returned by [PeerDiscovery] for the same node.
	Identity() string

	// Run participates in the election loop until ctx is cancelled.
	// onLeading is called each time this node becomes the leader, with a
	// sub-context that is cancelled when leadership is lost.
	// onLeading must return before the next election cycle may start.
	Run(ctx context.Context, onLeading func(ctx context.Context)) error
}

// --- Kubernetes Lease implementation ---

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

// LeaseElectorConfig holds configuration for Kubernetes Lease-based leader election.
type LeaseElectorConfig struct {
	Client        kubernetes.Interface
	Namespace     string
	LeaseName     string
	Identity      string // usually the pod name from the downward API
	LeaseDuration time.Duration
	RenewDeadline time.Duration
	RetryPeriod   time.Duration
}

// LeaseElector implements [Elector] using a Kubernetes Lease resource.
type LeaseElector struct {
	cfg LeaseElectorConfig
}

// NewLeaseElector creates a LeaseElector with the given config.
// Zero-value durations default to 15s/10s/2s (duration/renew/retry).
func NewLeaseElector(cfg LeaseElectorConfig) *LeaseElector {
	if cfg.LeaseDuration == 0 {
		cfg.LeaseDuration = defaultLeaseDuration
	}

	if cfg.RenewDeadline == 0 {
		cfg.RenewDeadline = defaultRenewDeadline
	}

	if cfg.RetryPeriod == 0 {
		cfg.RetryPeriod = defaultRetryPeriod
	}

	return &LeaseElector{cfg: cfg}
}

// Identity returns the pod name used as the Lease holder identity.
func (e *LeaseElector) Identity() string { return e.cfg.Identity }

// Run participates in the leader election loop until ctx is cancelled.
// onLeading is called with a sub-context when this node wins the lease.
func (e *LeaseElector) Run(ctx context.Context, onLeading func(ctx context.Context)) error {
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      e.cfg.LeaseName,
			Namespace: e.cfg.Namespace,
		},
		Client: e.cfg.Client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: e.cfg.Identity,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   e.cfg.LeaseDuration,
		RenewDeadline:   e.cfg.RenewDeadline,
		RetryPeriod:     e.cfg.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: onLeading,
			OnStoppedLeading: func() {},
			OnNewLeader:      func(_ string) {},
		},
	})

	return nil
}
