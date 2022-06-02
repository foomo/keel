package keeltemporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type (
	ActivityOption func(options *workflow.ActivityOptions)
)

func ActivityOptionsWithTaskQueue(v string) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.TaskQueue = v
	}
}

func ActivityOptionsWithScheduleToCloseTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.ScheduleToCloseTimeout = v
	}
}

func ActivityOptionsWithScheduleToStartTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.ScheduleToStartTimeout = v
	}
}

func ActivityOptionsWithStartToCloseTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.StartToCloseTimeout = v
	}
}

func ActivityOptionsWithHeartbeatTimeout(v time.Duration) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.HeartbeatTimeout = v
	}
}

func ActivityOptionsWithWaitForCancellation(v bool) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.WaitForCancellation = v
	}
}

func ActivityOptionsWithActivityID(v string) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.ActivityID = v
	}
}

func ActivityOptionsWithRetryPolicy(v *temporal.RetryPolicy) ActivityOption {
	return func(o *workflow.ActivityOptions) {
		o.RetryPolicy = v
	}
}

func WithActivityOptions(ctx workflow.Context, opts ...ActivityOption) workflow.Context {
	o := workflow.GetActivityOptions(ctx)
	for _, opt := range opts {
		opt(&o)
	}
	return workflow.WithActivityOptions(ctx, o)
}
