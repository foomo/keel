// Package leaderfollower provides a generic leader/follower coordination
// framework for distributed Go services running in Kubernetes.
//
// # Interfaces
//
// The framework is built from three pluggable interfaces:
//   - [Elector]: determines which node is the leader
//   - [PeerDiscovery]: resolves the current set of addressable peers
//   - [Protocol]: the coordination strategy executed by the leader
//
// Kubernetes implementations are included in this package:
//   - [LeaseElector]: leader election via Kubernetes Lease resource
//   - [PodDiscovery]: peer discovery via pod label selector
//
// For the batteries-included Kubernetes setup with three-phase commit,
// use [threephase.New] in the threephase sub-package, which also provides
// [threephase.ConfigMapCoordStore] for coordination state persistence.
//
// For manual wiring without the K8s defaults, use [NewCoordinator].
//
// # Protocol
//
// The threephase sub-package provides a three-phase commit protocol:
//   - Phase 1 (CanCommit): peers validate whether they can accept the proposed state
//   - Phase 2 (PreCommit): peers stage the state (default: no-op, degrades to 2PC)
//   - Phase 3 (DoCommit): peers apply the state atomically
//
// # ConfigMap layout
//
// The coordination ConfigMap managed by [threephase.ConfigMapCoordStore] uses the following keys:
//   - committed: JSON-encoded currently active state
//   - previous:  JSON-encoded previous state (for rollback after restart)
//   - proposed:  JSON-encoded next proposed state (set by external proposers)
//   - round:     JSON-encoded Round[S] for crash recovery (cleared after each round)
package leaderfollower
