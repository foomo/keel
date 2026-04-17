package threephase

import "time"

// Phase represents the current step in the three-phase commit protocol.
type Phase int

const (
	PhaseIdle          Phase = iota // no coordination in progress
	PhaseCanCommitting              // leader sent can_commit; awaiting peer votes
	PhasePreCommitting              // all peers agreed; leader sent pre_commit
	PhaseDoCommitting               // all peers pre-committed; leader sent do_commit
	PhaseAborting                   // at least one peer failed; leader sent abort
)

func (p Phase) String() string {
	switch p {
	case PhaseIdle:
		return "idle"
	case PhaseCanCommitting:
		return "can_committing"
	case PhasePreCommitting:
		return "pre_committing"
	case PhaseDoCommitting:
		return "do_committing"
	case PhaseAborting:
		return "aborting"
	default:
		return "unknown"
	}
}

// Round captures the full state of one coordination round.
// It is stored via [Store] so the leader can resume after a failover.
type Round[S any] struct {
	ID        string    `json:"id"`
	Phase     Phase     `json:"phase"`
	Proposed  S         `json:"proposed"`
	Previous  S         `json:"previous"` // state active before this round; used for rollback
	StartedAt time.Time `json:"startedAt"`
}
