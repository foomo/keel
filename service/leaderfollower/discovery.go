package leaderfollower

import (
	"context"
	"fmt"
	"net"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Peer represents a reachable node in the cluster.
type Peer struct {
	// ID is the stable identity of the peer. Must match [Elector.Identity] for
	// the same node so the coordinator can identify itself in the peer list.
	ID string

	// Addr is the base HTTP address of the peer's coordination server,
	Addr string
}

// PeerDiscovery resolves the current set of peers (including self).
// Implementations are called before each coordination round and must be
// safe for concurrent use.
type PeerDiscovery interface {
	Peers(ctx context.Context) ([]Peer, error)
}

// --- Kubernetes Pod implementation ---

// PodDiscoveryConfig holds configuration for Kubernetes pod-based peer discovery.
type PodDiscoveryConfig struct {
	Client        kubernetes.Interface
	Namespace     string
	LabelSelector string // e.g. "app.kubernetes.io/instance=myservice"
	CoordPort     int    // port the coord server listens on (e.g. 8090)
}

// PodDiscovery implements [PeerDiscovery] by listing running Kubernetes pods
// matching a label selector.
type PodDiscovery struct {
	cfg PodDiscoveryConfig
}

// NewPodDiscovery creates a PodDiscovery.
func NewPodDiscovery(cfg PodDiscoveryConfig) *PodDiscovery {
	return &PodDiscovery{cfg: cfg}
}

// Peers lists all running, non-terminating pods matching the label selector.
// The pod name is used as Peer.ID and the pod IP + CoordPort as Peer.Addr.
func (d *PodDiscovery) Peers(ctx context.Context) ([]Peer, error) {
	list, err := d.cfg.Client.CoreV1().Pods(d.cfg.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: d.cfg.LabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	var peers []Peer

	for _, p := range list.Items {
		if p.Status.Phase != corev1.PodRunning {
			continue
		}

		if p.DeletionTimestamp != nil {
			continue
		}

		if p.Status.PodIP == "" {
			continue
		}

		peers = append(peers, Peer{
			ID:   p.Name,
			Addr: "http://" + net.JoinHostPort(p.Status.PodIP, strconv.Itoa(d.cfg.CoordPort)),
		})
	}

	return peers, nil
}
