package threephase

import (
	"context"
	"time"
)

const (
	defaultCanCommitTimeout = 60 * time.Second
	defaultPreCommitTimeout = 30 * time.Second
	defaultDoCommitTimeout  = 120 * time.Second
	defaultPollInterval     = 2 * time.Second
)

// Option configures a [Protocol].
type Option[S any] func(*Protocol[S])

// WithPreCommit sets the callback called on every node when the leader issues
// pre_commit. Use this to stage the proposed state (e.g. load data into a
// staging buffer). Default: no-op, which degrades the protocol to effective 2PC.
func WithPreCommit[S any](fn func(ctx context.Context, proposed S) error) Option[S] {
	return func(p *Protocol[S]) { p.preCommit = fn }
}

// WithAbort sets the callback called on every node when the leader issues abort.
// Use this to discard any staged pre_commit state. Default: no-op.
func WithAbort[S any](fn func(ctx context.Context) error) Option[S] {
	return func(p *Protocol[S]) { p.abort = fn }
}

// WithRollback sets the callback called on every node when the leader needs to
// revert a failed do_commit. previous is the state to restore. Default: no-op.
func WithRollback[S any](fn func(ctx context.Context, previous S) error) Option[S] {
	return func(p *Protocol[S]) { p.rollback = fn }
}

// WithLeaderAfterCommit sets a callback invoked by the leader after all peers
// confirm do_commit. previous is the state before the round; committed is the
// newly active state.
//
// When using [New], the framework first writes committed and previous to the
// ConfigMap, then calls this hook. Use it for "run-once-globally" cleanup
// (deleting stale data, sending notifications, etc.).
// For per-pod logic, use the doCommit positional parameter instead.
func WithLeaderAfterCommit[S any](fn func(ctx context.Context, previous, committed S) error) Option[S] {
	return func(p *Protocol[S]) { p.afterCommit = fn }
}

// WithProposalEndpoint enables or disables the "propose" action on the coord
// HTTP server. When enabled, external services (e.g. batch jobs) can trigger a
// new commit round by POSTing {"action":"propose","payload":<JSON S>} to the
// coord endpoint.
//
// The coord URL for external callers is available via [CoordinatorHandle.ProposeURL].
func WithProposalEndpoint[S any](enabled bool) Option[S] {
	return func(p *Protocol[S]) { p.proposalEnabled = enabled }
}

// WithCanCommitTimeout sets how long the leader waits for all peers to vote
// on can_commit. Default: 60s.
func WithCanCommitTimeout[S any](d time.Duration) Option[S] {
	return func(p *Protocol[S]) { p.canCommitTimeout = d }
}

// WithPreCommitTimeout sets how long the leader waits for all peers to
// acknowledge pre_commit. Default: 30s.
func WithPreCommitTimeout[S any](d time.Duration) Option[S] {
	return func(p *Protocol[S]) { p.preCommitTimeout = d }
}

// WithDoCommitTimeout sets how long the leader waits for all peers to confirm
// do_commit. Default: 120s.
func WithDoCommitTimeout[S any](d time.Duration) Option[S] {
	return func(p *Protocol[S]) { p.doCommitTimeout = d }
}

// WithPollInterval sets the retry interval for re-polling peers during any phase.
// Default: 2s.
func WithPollInterval[S any](d time.Duration) Option[S] {
	return func(p *Protocol[S]) { p.pollInterval = d }
}
