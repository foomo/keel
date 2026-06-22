package keel

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/foomo/gofuncy"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/env"
	internalotel "github.com/foomo/keel/internal/otel"
	"github.com/foomo/keel/log"
	keelsemconv "github.com/foomo/keel/semconv"
	"github.com/foomo/keel/telemetry"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// StepFn is a single unit of work within a Job. It receives the (cancellable)
	// job context and an enriched logger and returns an error to fail the job.
	StepFn func(ctx context.Context, l *zap.Logger) error

	jobStep struct {
		name string
		fn   StepFn
	}

	// jobPusher pushes/flushes telemetry during job finalization.
	jobPusher func(ctx context.Context) error
)

// Job runs a keel workload to completion, as suited for a Kubernetes Job.
//
// Unlike Server (which blocks until a shutdown signal), a Job runs its steps in
// order, then exits: RunE returns the step error (or nil), and Run translates that
// into a process exit code so Kubernetes can apply backoffLimit. SIGINT/SIGTERM are
// treated as an abnormal interruption that cancels the running step. No init HTTP
// services are started; metrics are pushed/flushed on exit rather than exposed.
type Job struct {
	name            string
	steps           []jobStep
	parallel        bool
	parallelLimit   int
	meterProvider   metric.MeterProvider
	traceProvider   trace.TracerProvider
	loggerProvider  otellog.LoggerProvider
	pushers         []jobPusher
	shutdownSignals []os.Signal
	gracefulPeriod  time.Duration
	timeout         time.Duration
	syncClosers     []any
	ctx             context.Context
	l               *zap.Logger
	c               *viper.Viper
}

// NewJob creates a new Job with the given options. The job name defaults to the
// OTEL_SERVICE_NAME (falling back to telemetry.DefaultServiceName) and can be
// overridden with JobWithName. It is used as the root routine/span label, the
// Pushgateway group, and a log field.
func NewJob(opts ...JobOption) *Job {
	inst := &Job{
		gracefulPeriod:  time.Duration(env.GetInt("KEEL_GRACEFUL_PERIOD", 30)) * time.Second,
		shutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		ctx:             context.Background(),
		c:               config.Config(),
		l:               log.Logger(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	if inst.name == "" {
		inst.name = env.Get("OTEL_SERVICE_NAME", telemetry.DefaultServiceName)
	}

	inst.l = log.WithAttributes(inst.l, keelsemconv.KeelServiceType("job"), keelsemconv.KeelServiceName(inst.name))

	{ // setup telemetry
		otel.SetLogger(logr.New(internalotel.NewLogger(inst.l)))
		otel.SetErrorHandler(internalotel.NewErrorHandler(inst.l))
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		if inst.meterProvider == nil {
			inst.meterProvider = telemetry.NewNoopMeterProvider()
		}

		if inst.traceProvider == nil {
			inst.traceProvider = telemetry.NewNoopTraceProvider()
		}

		if inst.loggerProvider == nil {
			inst.loggerProvider = telemetry.NewNoopLoggerProvider()
		} else {
			inst.l = inst.l.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewTee(c, telemetry.NewZapBridgeCore(inst.loggerProvider))
			}))
		}
	}

	return inst
}

// Logger returns the job logger.
func (j *Job) Logger() *zap.Logger {
	return j.l
}

// Config returns the job config.
func (j *Job) Config() *viper.Viper {
	return j.c
}

// Context returns the job context.
func (j *Job) Context() context.Context {
	return j.ctx
}

// Meter returns the job meter.
func (j *Job) Meter() metric.Meter {
	return telemetry.Meter()
}

// Tracer returns the job tracer.
func (j *Job) Tracer() trace.Tracer {
	return telemetry.Tracer()
}

// AddStep adds a step to be run in registration order.
func (j *Job) AddStep(name string, fn StepFn) {
	j.steps = append(j.steps, jobStep{name: name, fn: fn})
}

// AddCloser registers a closer to be called during job finalization.
func (j *Job) AddCloser(closer any) {
	if !IsCloser(closer) {
		j.l.Warn("unable to add closer", log.FValue(fmt.Sprintf("%T", closer)))
		return
	}

	if slices.Contains(j.syncClosers, closer) {
		return
	}

	j.syncClosers = append(j.syncClosers, closer)
}

// AddClosers registers the given closers to be called during job finalization.
func (j *Job) AddClosers(closers ...any) {
	for _, closer := range closers {
		j.AddCloser(closer)
	}
}

// Run executes the job and exits the process: 0 on success, 1 on failure. This is
// the convenience entrypoint for a job's main(); use RunE if you need to handle the
// error yourself.
func (j *Job) Run() {
	if err := j.RunE(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

// RunE executes the job steps in order and finalizes telemetry and resources. It
// returns the first step error (or the context error on interruption/timeout), or
// nil on success. Finalization (metric push/flush, closers) always runs, even on
// error or interruption.
func (j *Job) RunE() error {
	j.l.With(log.Attributes(telemetry.EnvAttributes()...)...).Info("starting keel job")

	start := time.Now()

	ctx, stop := signal.NotifyContext(j.ctx, j.shutdownSignals...)
	defer stop()

	if j.timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, j.timeout)
		defer cancel()
	}

	// always finalize, even on error or interruption
	defer j.finalize()

	err := j.run(ctx)
	if err != nil {
		log.WithError(j.l, err).Error("keel job failed", log.FDuration(time.Since(start)))
	} else {
		j.l.Info("keel job completed", log.FDuration(time.Since(start)))
	}

	return err
}

// run executes the steps as the root job routine. gofuncy provides the span,
// panic recovery, and goroutine metrics; steps route through the job's providers.
func (j *Job) run(ctx context.Context) error {
	err := gofuncy.Do(ctx, j.runSteps,
		gofuncy.WithName("job "+j.name),
		gofuncy.WithTracerProvider(j.traceProvider),
		gofuncy.WithMeterProvider(j.meterProvider),
	)

	// surface interruption/timeout even if no step reported it
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return err
}

// runSteps runs the steps sequentially (default) or concurrently (JobWithParallel).
func (j *Job) runSteps(ctx context.Context) error {
	if j.parallel {
		g := gofuncy.NewGroup(ctx,
			gofuncy.WithName("steps"),
			gofuncy.WithLimit(j.parallelLimit),
			gofuncy.WithFailFast(),
			gofuncy.WithTracerProvider(j.traceProvider),
			gofuncy.WithMeterProvider(j.meterProvider),
		)

		for _, step := range j.steps {
			g.Add(func(ctx context.Context) error {
				return j.execStep(ctx, step)
			}, gofuncy.WithName("step "+step.name))
		}

		return g.Wait()
	}

	for _, step := range j.steps {
		if err := gofuncy.Do(ctx,
			func(ctx context.Context) error { return j.execStep(ctx, step) },
			gofuncy.WithName("step "+step.name),
			gofuncy.WithTracerProvider(j.traceProvider),
			gofuncy.WithMeterProvider(j.meterProvider),
		); err != nil {
			return err
		}
	}

	return nil
}

// execStep adapts a StepFn to gofuncy's Func, adding the per-step logger and
// completion logging. The span, panic recovery, and metrics are gofuncy's.
func (j *Job) execStep(ctx context.Context, step jobStep) error {
	l := j.l.With(zap.String("keel_job_step", step.name))
	l.Info("starting keel job step")

	start := time.Now()

	if err := step.fn(ctx, l); err != nil {
		log.WithError(l, err).Error("keel job step failed", log.FDuration(time.Since(start)))
		return err
	}

	l.Info("keel job step completed", log.FDuration(time.Since(start)))

	return nil
}

// finalize pushes/flushes telemetry and closes registered resources within the
// graceful period. It uses a context detached from cancellation so cleanup still
// runs after an interruption or timeout.
func (j *Job) finalize() {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(j.ctx), j.gracefulPeriod)
	defer cancel()

	j.l.Info("keel job finalize", zap.Duration("graceful_period", j.gracefulPeriod))

	for _, push := range j.pushers {
		if err := push(ctx); err != nil {
			log.WithError(j.l, err).Warn("keel job finalize: push failed")
		}
	}

	closers := append(slices.Clone(j.syncClosers), j.traceProvider, j.meterProvider, j.loggerProvider)
	closeAll(ctx, j.l, closers)

	j.l.Info("keel job finalize: complete")
}
