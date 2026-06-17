package keel

import "context"

// AddPusherForTest exposes the unexported pusher hook so tests can assert that
// finalization runs pushers (including on error/interruption) without relying on
// the global Prometheus registry.
func (j *Job) AddPusherForTest(fn func(ctx context.Context) error) {
	j.pushers = append(j.pushers, jobPusher(fn))
}

// NameForTest exposes the resolved job name for assertions.
func (j *Job) NameForTest() string {
	return j.name
}
