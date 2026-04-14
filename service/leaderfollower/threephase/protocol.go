package threephase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	lf "github.com/foomo/keel/service/leaderfollower"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// actions used in FanOut / Handle dispatch.
const (
	actionCanCommit = "can_commit" // phase 1: can you accept this proposed state?
	actionPreCommit = "pre_commit" // phase 2: stage the state (no-op default)
	actionDoCommit  = "do_commit"  // phase 3: apply the state
	actionAbort     = "abort"      // abort the current round
	actionRollback  = "rollback"   // revert to previous state after a failed do_commit
	actionPropose   = "propose"    // set a proposed state (only when proposalEnabled)
)

// peerVote is the response body for can_commit.
type peerVote struct {
	Ready bool   `json:"ready"`
	Error string `json:"error,omitempty"`
}

// peerAck is the response body for pre_commit, do_commit, abort, and rollback.
type peerAck struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// rollbackPayload carries the previous state sent during a rollback fan-out.
type rollbackPayload[S any] struct {
	Previous S `json:"previous"`
}

// Protocol implements [lf.Protocol] using a three-phase commit strategy.
//
//   - Phase 1 (can_commit): leader asks all peers "can you accept this state?"
//   - Phase 2 (pre_commit): leader instructs peers to stage the state (no-op default)
//   - Phase 3 (do_commit): leader instructs peers to apply the state
//
// If PreCommit is not set, the protocol degrades to effective two-phase commit.
// S must be JSON-serialisable.
//
// canCommit and doCommit are required. Use Option functions for the rest.
type Protocol[S any] struct {
	l     *zap.Logger
	store CoordStore[S]

	// Required callbacks
	canCommit func(ctx context.Context, proposed S) error
	doCommit  func(ctx context.Context, proposed S) error

	// Optional callbacks (all default to no-op).
	preCommit   func(ctx context.Context, proposed S) error
	abort       func(ctx context.Context) error
	rollback    func(ctx context.Context, previous S) error
	afterCommit func(ctx context.Context, previous, committed S) error

	proposalEnabled bool // when true, Handle accepts the "propose" action

	canCommitTimeout time.Duration
	preCommitTimeout time.Duration
	doCommitTimeout  time.Duration
	pollInterval     time.Duration
}

// NewProtocol creates a Protocol with required canCommit and doCommit callbacks.
// Configure optional callbacks and timeouts via Option functions.
//
// For the batteries-included Kubernetes setup, use [New] in new.go instead.
func NewProtocol[S any](
	l *zap.Logger,
	store CoordStore[S],
	canCommit func(ctx context.Context, proposed S) error,
	doCommit func(ctx context.Context, proposed S) error,
	opts ...Option[S],
) *Protocol[S] {
	p := &Protocol[S]{
		l:                l,
		store:            store,
		canCommit:        canCommit,
		doCommit:         doCommit,
		canCommitTimeout: defaultCanCommitTimeout,
		preCommitTimeout: defaultPreCommitTimeout,
		doCommitTimeout:  defaultDoCommitTimeout,
		pollInterval:     defaultPollInterval,
	}
	for _, o := range opts {
		o(p)
	}

	return p
}

// Lead implements [lf.Protocol]. Called on the leader for its leadership term.
func (p *Protocol[S]) Lead(ctx context.Context, fanOut lf.FanOut) error {
	// Load the last committed state once on entry so we can:
	//   1. filter out already-committed proposals in WatchProposed
	//   2. use it as Previous in each new round
	lastCommitted, _, _, err := p.store.LoadCommitted(ctx)
	if err != nil {
		p.l.Warn("threephase: failed to load committed state on leader start", zap.Error(err))
		// Continue with zero value — not fatal.
	}

	// Resume any in-flight round from a previous leadership term.
	if round := p.loadRound(ctx); round != nil {
		p.l.Info("threephase: resuming in-flight round after re-election",
			zap.String("round", round.ID), zap.String("phase", round.Phase.String()))
		p.resumeRound(ctx, fanOut, round)
		// After resume, the committed state may have changed — reload it.
		if committed, _, ok, loadErr := p.store.LoadCommitted(ctx); ok && loadErr == nil {
			lastCommitted = committed
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		proposed, err := p.store.WatchProposed(ctx, lastCommitted)
		if err != nil {
			if ctx.Err() != nil {
				return nil //nolint:nilerr // context cancellation is a clean shutdown, not an error
			}

			p.l.Error("threephase: WatchProposed error", zap.Error(err))

			continue
		}

		round := &Round[S]{
			ID:        uuid.New().String(),
			Phase:     PhaseCanCommitting,
			Proposed:  proposed,
			Previous:  lastCommitted,
			StartedAt: time.Now(),
		}
		p.saveRound(ctx, round)
		p.coordinateRound(ctx, fanOut, round)

		// Update lastCommitted only after a successful round.
		if committed, _, ok, loadErr := p.store.LoadCommitted(ctx); ok && loadErr == nil {
			lastCommitted = committed
		}
	}
}

// resumeRound handles in-flight rounds found after re-election.
func (p *Protocol[S]) resumeRound(ctx context.Context, fanOut lf.FanOut, round *Round[S]) {
	switch round.Phase {
	case PhaseCanCommitting, PhasePreCommitting:
		// Safe to abort: no committed state yet.
		p.l.Info("threephase: aborting unfinished round (pre-commit phase, safe)",
			zap.String("round", round.ID), zap.String("phase", round.Phase.String()))
		p.fanOutIgnoreErrors(ctx, fanOut, actionAbort, nil)
		p.clearRound(ctx)
	case PhaseDoCommitting:
		// Partial commit possible — roll back if a handler is registered.
		p.l.Warn("threephase: unfinished do_commit round after re-election",
			zap.String("round", round.ID))

		if p.rollback != nil {
			prevPayload, err := json.Marshal(rollbackPayload[S]{Previous: round.Previous})
			if err != nil {
				p.l.Error("threephase: failed to marshal rollback payload — peers cannot roll back",
					zap.String("round", round.ID), zap.Error(err))
				p.clearRound(ctx)

				return
			}

			p.fanOutIgnoreErrors(ctx, fanOut, actionRollback, prevPayload)
		}

		p.clearRound(ctx)
	default:
		p.clearRound(ctx)
	}
}

// coordinateRound drives one complete three-phase round.
func (p *Protocol[S]) coordinateRound(ctx context.Context, fanOut lf.FanOut, round *Round[S]) {
	p.l.Info("threephase: starting round", zap.String("round", round.ID))

	proposedPayload, err := json.Marshal(round.Proposed)
	if err != nil {
		p.l.Error("threephase: failed to marshal proposed state — aborting round",
			zap.String("round", round.ID), zap.Error(err))
		p.clearRound(ctx)

		return
	}

	// Phase 1: can_commit — wait until all peers are ready or give up.
	if !p.waitForAll(ctx, fanOut, actionCanCommit, proposedPayload, p.canCommitTimeout, isVoteReady) {
		p.l.Warn("threephase: aborting round at can_commit", zap.String("round", round.ID))
		p.fanOutIgnoreErrors(ctx, fanOut, actionAbort, nil)
		p.clearRound(ctx)

		return
	}

	// Phase 2: pre_commit — all peers stage the state.
	round.Phase = PhasePreCommitting
	p.saveRound(ctx, round)

	if !p.waitForAll(ctx, fanOut, actionPreCommit, proposedPayload, p.preCommitTimeout, isAckOK) {
		p.l.Warn("threephase: aborting round at pre_commit", zap.String("round", round.ID))
		p.fanOutIgnoreErrors(ctx, fanOut, actionAbort, nil)
		p.clearRound(ctx)

		return
	}

	// Phase 3: do_commit — all peers apply the state.
	round.Phase = PhaseDoCommitting
	p.saveRound(ctx, round)
	p.l.Info("threephase: all peers pre-committed — issuing do_commit", zap.String("round", round.ID))

	if !p.waitForAll(ctx, fanOut, actionDoCommit, proposedPayload, p.doCommitTimeout, isAckOK) {
		p.l.Warn("threephase: do_commit failed — rolling back", zap.String("round", round.ID))

		if p.rollback != nil {
			prevPayload, err := json.Marshal(rollbackPayload[S]{Previous: round.Previous})
			if err != nil {
				p.l.Error("threephase: failed to marshal rollback payload — peers cannot roll back",
					zap.String("round", round.ID), zap.Error(err))
				p.clearRound(ctx)

				return
			}

			p.fanOutIgnoreErrors(ctx, fanOut, actionRollback, prevPayload)
		}

		p.clearRound(ctx)

		return
	}

	p.l.Info("threephase: round complete", zap.String("round", round.ID))
	p.clearRound(ctx)

	if p.afterCommit != nil {
		if err := p.afterCommit(ctx, round.Previous, round.Proposed); err != nil {
			p.l.Error("threephase: afterCommit hook failed", zap.String("round", round.ID), zap.Error(err))
		}
	}
}

// waitForAll retries action until all peers satisfy check or timeout/ctx cancel.
// checker returns (satisfied bool, hardFail bool) — hardFail triggers immediate abort.
func (p *Protocol[S]) waitForAll(
	ctx context.Context,
	fanOut lf.FanOut,
	action string,
	payload []byte,
	timeout time.Duration,
	checker func(body []byte) (ok bool, hardFail bool),
) bool {
	deadline := time.Now().Add(timeout)

	for {
		if ctx.Err() != nil {
			return false
		}

		results := fanOut.All(ctx, lf.FanOutRequest{Action: action, Payload: payload})
		allOK := true

		for _, r := range results {
			if r.Err != nil {
				p.l.Warn("threephase: peer unreachable, retrying",
					zap.String("action", action), zap.String("peer", r.Peer.ID), zap.Error(r.Err))

				allOK = false

				continue
			}

			ok, hard := checker(r.Body)
			if !ok {
				if hard {
					p.l.Warn("threephase: peer hard-failed — aborting",
						zap.String("action", action), zap.String("peer", r.Peer.ID))

					return false
				}

				allOK = false
			}
		}

		if allOK && len(results) > 0 {
			return true
		}

		if time.Now().After(deadline) {
			p.l.Warn("threephase: timeout", zap.String("action", action))
			return false
		}

		select {
		case <-ctx.Done():
			return false
		case <-time.After(p.pollInterval):
		}
	}
}

// fanOutIgnoreErrors fires action to all peers, logging but not failing on errors.
func (p *Protocol[S]) fanOutIgnoreErrors(ctx context.Context, fanOut lf.FanOut, action string, payload []byte) {
	for _, r := range fanOut.All(ctx, lf.FanOutRequest{Action: action, Payload: payload}) {
		if r.Err != nil {
			p.l.Warn("threephase: fan-out error",
				zap.String("action", action), zap.String("peer", r.Peer.ID), zap.Error(r.Err))
		}
	}
}

// Handle implements [lf.Protocol]. Called on every node for incoming /coord POSTs.
func (p *Protocol[S]) Handle(ctx context.Context, w http.ResponseWriter, action string, body []byte) error {
	switch action {
	case actionCanCommit:
		return p.handleCanCommit(ctx, w, body)
	case actionPreCommit:
		return p.handlePreCommit(ctx, w, body)
	case actionDoCommit:
		return p.handleDoCommit(ctx, w, body)
	case actionAbort:
		return p.handleAbort(ctx, w)
	case actionRollback:
		return p.handleRollback(ctx, w, body)
	case actionPropose:
		if p.proposalEnabled {
			return p.handlePropose(ctx, w, body)
		}

		http.Error(w, "propose endpoint not enabled", http.StatusForbidden)

		return nil
	default:
		http.Error(w, fmt.Sprintf("unknown action %q", action), http.StatusBadRequest)
		return nil
	}
}

func (p *Protocol[S]) handleCanCommit(ctx context.Context, w http.ResponseWriter, body []byte) error {
	var proposed S
	if err := json.Unmarshal(body, &proposed); err != nil {
		http.Error(w, "invalid can_commit payload: "+err.Error(), http.StatusBadRequest)
		return nil //nolint:nilerr // error written to HTTP response; nothing to propagate to caller
	}

	if err := p.canCommit(ctx, proposed); err != nil {
		p.l.Warn("threephase: can_commit validation failed", zap.Error(err))
		writeJSON(w, http.StatusOK, peerVote{Ready: false, Error: err.Error()})

		return nil
	}

	writeJSON(w, http.StatusOK, peerVote{Ready: true})

	return nil
}

func (p *Protocol[S]) handlePreCommit(ctx context.Context, w http.ResponseWriter, body []byte) error {
	var proposed S
	if err := json.Unmarshal(body, &proposed); err != nil {
		http.Error(w, "invalid pre_commit payload: "+err.Error(), http.StatusBadRequest)
		return nil //nolint:nilerr // error written to HTTP response; nothing to propagate to caller
	}

	if p.preCommit != nil {
		if err := p.preCommit(ctx, proposed); err != nil {
			p.l.Error("threephase: pre_commit failed", zap.Error(err))
			writeJSON(w, http.StatusOK, peerAck{OK: false, Error: err.Error()})

			return nil
		}
	}

	writeJSON(w, http.StatusOK, peerAck{OK: true})

	return nil
}

func (p *Protocol[S]) handleDoCommit(ctx context.Context, w http.ResponseWriter, body []byte) error {
	var proposed S
	if err := json.Unmarshal(body, &proposed); err != nil {
		http.Error(w, "invalid do_commit payload: "+err.Error(), http.StatusBadRequest)
		return nil //nolint:nilerr // error written to HTTP response; nothing to propagate to caller
	}

	if err := p.doCommit(ctx, proposed); err != nil {
		p.l.Error("threephase: do_commit failed", zap.Error(err))
		writeJSON(w, http.StatusOK, peerAck{OK: false, Error: err.Error()})

		return nil
	}

	writeJSON(w, http.StatusOK, peerAck{OK: true})

	return nil
}

func (p *Protocol[S]) handleAbort(ctx context.Context, w http.ResponseWriter) error {
	if p.abort != nil {
		if err := p.abort(ctx); err != nil {
			p.l.Error("threephase: abort handler error", zap.Error(err))
		}
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (p *Protocol[S]) handleRollback(ctx context.Context, w http.ResponseWriter, body []byte) error {
	var rp rollbackPayload[S]
	if err := json.Unmarshal(body, &rp); err != nil {
		http.Error(w, "invalid rollback payload: "+err.Error(), http.StatusBadRequest)
		return nil //nolint:nilerr // error written to HTTP response; nothing to propagate to caller
	}

	if p.rollback != nil {
		if err := p.rollback(ctx, rp.Previous); err != nil {
			p.l.Error("threephase: rollback failed", zap.Error(err))
			writeJSON(w, http.StatusOK, peerAck{OK: false, Error: err.Error()})

			return nil
		}
	}

	writeJSON(w, http.StatusOK, peerAck{OK: true})

	return nil
}

// handlePropose is called when a client POSTs action="propose" to the coord server.
// It deserialises the payload as S and writes it to the store as the next proposed state.
func (p *Protocol[S]) handlePropose(ctx context.Context, w http.ResponseWriter, body []byte) error {
	var proposed S
	if err := json.Unmarshal(body, &proposed); err != nil {
		http.Error(w, "invalid propose payload: "+err.Error(), http.StatusBadRequest)
		return nil //nolint:nilerr // error written to HTTP response; nothing to propagate to caller
	}

	if err := p.store.SetProposed(ctx, proposed); err != nil {
		p.l.Error("threephase: SetProposed failed", zap.Error(err))
		http.Error(w, "failed to store proposed state: "+err.Error(), http.StatusInternalServerError)

		return nil
	}

	w.WriteHeader(http.StatusAccepted)

	return nil
}

// --- store helpers ---

func (p *Protocol[S]) loadRound(ctx context.Context) *Round[S] {
	if p.store == nil {
		return nil
	}

	data, ok, err := p.store.LoadRound(ctx)
	if err != nil {
		p.l.Warn("threephase: failed to load round from store", zap.Error(err))

		return nil
	}

	if !ok {
		return nil
	}

	var round Round[S]
	if err := json.Unmarshal(data, &round); err != nil {
		p.l.Warn("threephase: failed to unmarshal round from store", zap.Error(err))

		return nil
	}

	return &round
}

func (p *Protocol[S]) saveRound(ctx context.Context, round *Round[S]) {
	if p.store == nil {
		return
	}

	data, err := json.Marshal(round)
	if err != nil {
		p.l.Warn("threephase: failed to marshal round for store", zap.Error(err))
		return
	}

	if err := p.store.SaveRound(ctx, data); err != nil {
		p.l.Warn("threephase: failed to save round to store", zap.Error(err))
	}
}

func (p *Protocol[S]) clearRound(ctx context.Context) {
	if p.store == nil {
		return
	}

	if err := p.store.ClearRound(ctx); err != nil {
		p.l.Warn("threephase: failed to clear round from store", zap.Error(err))
	}
}

// --- response helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) //nolint:errchkjson // v is always a concrete serializable struct (peerVote/peerAck)
}

// isVoteReady checks a can_commit response body.
// Returns (ready, hardFail): hardFail=true if the peer explicitly rejected.
func isVoteReady(body []byte) (bool, bool) {
	var v peerVote

	_ = json.Unmarshal(body, &v)
	if !v.Ready && v.Error != "" {
		return false, true // explicit rejection
	}

	return v.Ready, false
}

// isAckOK checks a pre_commit/do_commit/rollback response body.
// Returns (ok, hardFail): hardFail=true if the peer reported an explicit error.
func isAckOK(body []byte) (bool, bool) {
	var a peerAck

	_ = json.Unmarshal(body, &a)
	if !a.OK && a.Error != "" {
		return false, true
	}

	return a.OK, false
}
