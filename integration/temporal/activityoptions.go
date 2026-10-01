package keeltemporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type (
	// ActivityOption modifies [workflow.ActivityOptions]; see [WithActivityOptions].
	ActivityOption func(options *workflow.ActivityOptions)
)

// ActivityOptionsWithTaskQueue sets the task queue the activity is scheduled on.
func ActivityOptionsWithTaskQueue(v string) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.TaskQueue = v
	}
}

// ActivityOptionsWithScheduleToCloseTimeout sets the total time allowed for
// the activity, including retries.
func ActivityOptionsWithScheduleToCloseTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.ScheduleToCloseTimeout = v
	}
}

// ActivityOptionsWithScheduleToStartTimeout sets the time an activity task may
// wait in the task queue before being picked up by a worker.
func ActivityOptionsWithScheduleToStartTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.ScheduleToStartTimeout = v
	}
}

// ActivityOptionsWithStartToCloseTimeout sets the maximum duration of a single
// activity attempt.
func ActivityOptionsWithStartToCloseTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.StartToCloseTimeout = v
	}
}

// ActivityOptionsWithHeartbeatTimeout sets the maximum time allowed between
// activity heartbeats.
func ActivityOptionsWithHeartbeatTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.HeartbeatTimeout = v
	}
}

// ActivityOptionsWithWaitForCancellation sets whether the workflow waits for
// the activity to complete its cancellation before proceeding.
func ActivityOptionsWithWaitForCancellation(v bool) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.WaitForCancellation = v
	}
}

// ActivityOptionsWithActivityID sets the business level activity ID.
func ActivityOptionsWithActivityID(v string) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.ActivityID = v
	}
}

// ActivityOptionsWithRetryPolicy sets the retry policy of the activity.
func ActivityOptionsWithRetryPolicy(v *temporal.RetryPolicy) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.RetryPolicy = v
	}
}

// WithActivityOptions returns a copy of ctx whose activity options are the
// options already present in ctx modified by opts.
func WithActivityOptions(ctx workflow.Context, opts ...ActivityOption) workflow.Context {
	o := workflow.GetActivityOptions(ctx)
	for _, opt := range opts {
		opt(&o)
	}

	return workflow.WithActivityOptions(ctx, o)
}
