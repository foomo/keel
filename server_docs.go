//go:build docs
// +build docs

package keel

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"time"

	markdowntable "github.com/fbiville/markdown-table-formatter/pkg/markdown"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/service"
	"github.com/foomo/keel/telemetry/nonrecording"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	otelglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Server struct
type Server struct {
	services        []Service
	initServices    []Service
	meter           metric.Meter
	meterProvider   metric.MeterProvider
	tracer          trace.Tracer
	traceProvider   trace.TracerProvider
	shutdownSignals []os.Signal
	shutdownTimeout time.Duration
	closers         []interface{}
	probes          map[healthz.Type][]interface{}
	ctx             context.Context
	gCtx            context.Context
	l               *zap.Logger
	c               *viper.Viper
}

func NewServer(opts ...Option) *Server {
	inst := &Server{
		probes:        map[healthz.Type][]interface{}{},
		meterProvider: nonrecording.NewNoopMeterProvider(),
		traceProvider: trace.NewNoopTracerProvider(),
		ctx:           context.Background(),
		c:             config.Config(),
		l:             log.Logger(),
	}

	inst.meter = inst.meterProvider.Meter("")
	otelglobal.SetMeterProvider(inst.meterProvider)
	inst.tracer = inst.traceProvider.Tracer("")
	otel.SetTracerProvider(inst.traceProvider)

	// add probe
	inst.AddAlwaysHealthzers(inst)

	return inst
}

// Logger returns server logger
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Meter returns the implementation meter
func (s *Server) Meter() metric.Meter {
	return s.meter
}

// Tracer returns the implementation tracer
func (s *Server) Tracer() trace.Tracer {
	return s.tracer
}

// Config returns server config
func (s *Server) Config() *viper.Viper {
	return s.c
}

// Context returns server context
func (s *Server) Context() context.Context {
	return s.ctx
}

// CancelContext returns server's cancel context
func (s *Server) CancelContext() context.Context {
	return s.ctx
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	for _, value := range s.services {
		if value == service {
			return
		}
	}
	s.services = append(s.services, service)
	s.AddAlwaysHealthzers(service)
	s.AddCloser(service)
}

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddCloser adds a closer to be called on shutdown
func (s *Server) AddCloser(closer interface{}) {
	for _, value := range s.closers {
		if value == closer {
			return
		}
	}
	switch closer.(type) {
	case interfaces.Closer,
		interfaces.ErrorCloser,
		interfaces.CloserWithContext,
		interfaces.ErrorCloserWithContext,
		interfaces.Shutdowner,
		interfaces.ErrorShutdowner,
		interfaces.ShutdownerWithContext,
		interfaces.ErrorShutdownerWithContext,
		interfaces.Stopper,
		interfaces.ErrorStopper,
		interfaces.StopperWithContext,
		interfaces.ErrorStopperWithContext,
		interfaces.Unsubscriber,
		interfaces.ErrorUnsubscriber,
		interfaces.UnsubscriberWithContext,
		interfaces.ErrorUnsubscriberWithContext:
		s.closers = append(s.closers, closer)
	default:
		s.l.Warn("unable to add closer", log.FValue(fmt.Sprintf("%T", closer)))
	}
}

// AddClosers adds the given closers to be called on shutdown
func (s *Server) AddClosers(closers ...interface{}) {
	for _, closer := range closers {
		s.AddCloser(closer)
	}
}

// AddHealthzer adds a probe to be called on healthz checks
func (s *Server) AddHealthzer(typ healthz.Type, probe interface{}) {
	switch probe.(type) {
	case healthz.BoolHealthzer,
		healthz.BoolHealthzerWithContext,
		healthz.ErrorHealthzer,
		healthz.ErrorHealthzWithContext,
		interfaces.ErrorPinger,
		interfaces.ErrorPingerWithContext:
		s.probes[typ] = append(s.probes[typ], probe)
	default:
		s.l.Debug("not a healthz probe", log.FValue(fmt.Sprintf("%T", probe)))
	}
}

// AddHealthzers adds the given probes to be called on healthz checks
func (s *Server) AddHealthzers(typ healthz.Type, probes ...interface{}) {
	for _, probe := range probes {
		s.AddHealthzer(typ, probe)
	}
}

// AddAlwaysHealthzers adds the probes to be called on any healthz checks
func (s *Server) AddAlwaysHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeAlways, probes...)
}

// AddStartupHealthzers adds the startup probes to be called on healthz checks
func (s *Server) AddStartupHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeStartup, probes...)
}

// AddLivenessHealthzers adds the liveness probes to be called on healthz checks
func (s *Server) AddLivenessHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeLiveness, probes...)
}

// AddReadinessHealthzers adds the readiness probes to be called on healthz checks
func (s *Server) AddReadinessHealthzers(probes ...interface{}) {
	s.AddHealthzers(healthz.TypeReadiness, probes...)
}

// IsCanceled returns true if the internal errgroup has been canceled
func (s *Server) IsCanceled() bool {
	return errors.Is(s.gCtx.Err(), context.Canceled)
}

// Healthz returns true if the server is running
func (s *Server) Healthz() error {
	return nil
}

// Run runs the server
func (s *Server) Run() {
	// add init services to closers
	for _, initService := range s.initServices {
		s.AddClosers(initService)
	}

	md := &MD{}

	{
		var rows [][]string
		for _, key := range s.Config().AllKeys() {
			rows = append(rows, []string{
				code(key),
				code(s.Config().GetString(key)),
			})
		}
		if len(rows) > 0 {
			md.Println("## Config")
			md.Println("")
			md.Println("List of all registered config variabled with their defaults.")
			md.Println("")
			md.Table([]string{"Key", "Default"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		for _, value := range s.initServices {
			if v, ok := value.(*service.HTTP); ok {
				t := reflect.TypeOf(v)
				rows = append(rows, []string{
					code(v.Name()),
					code(t.String()),
					stringer(v),
				})
			}
		}
		if len(rows) > 0 {
			md.Println("## Init Services")
			md.Println("")
			md.Println("List of all registerd init services that are being immediately started.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Address"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		for _, value := range s.services {
			if v, ok := value.(*service.HTTP); ok {
				t := reflect.TypeOf(v)
				rows = append(rows, []string{
					code(v.Name()),
					code(t.String()),
					stringer(v),
				})
			}
		}
		if len(rows) > 0 {
			md.Println("## Services")
			md.Println("")
			md.Println("List of all registered services that are being started.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Description"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		for k, probes := range s.probes {
			for _, probe := range probes {
				t := reflect.TypeOf(probe)
				rows = append(rows, []string{
					code(k.String()),
					code(t.String()),
				})
			}
		}
		if len(rows) > 0 {
			md.Println("## Health probes")
			md.Println("")
			md.Println("List of all registered healthz probes that are being called during startup and runntime.")
			md.Println("")
			md.Table([]string{"Name", "Type"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		for _, value := range s.closers {
			t := reflect.TypeOf(value)
			var closer string
			switch value.(type) {
			case interfaces.Closer:
				closer = "Closer"
			case interfaces.ErrorCloser:
				closer = "ErrorCloser"
			case interfaces.CloserWithContext:
				closer = "CloserWithContext"
			case interfaces.ErrorCloserWithContext:
				closer = "ErrorCloserWithContext"
			case interfaces.Shutdowner:
				closer = "Shutdowner"
			case interfaces.ErrorShutdowner:
				closer = "ErrorShutdowner"
			case interfaces.ShutdownerWithContext:
				closer = "ShutdownerWithContext"
			case interfaces.ErrorShutdownerWithContext:
				closer = "ErrorShutdownerWithContext"
			case interfaces.Stopper:
				closer = "Stopper"
			case interfaces.ErrorStopper:
				closer = "ErrorStopper"
			case interfaces.StopperWithContext:
				closer = "StopperWithContext"
			case interfaces.ErrorStopperWithContext:
				closer = "ErrorStopperWithContext"
			case interfaces.Unsubscriber:
				closer = "Unsubscriber"
			case interfaces.ErrorUnsubscriber:
				closer = "ErrorUnsubscriber"
			case interfaces.UnsubscriberWithContext:
				closer = "UnsubscriberWithContext"
			case interfaces.ErrorUnsubscriberWithContext:
				closer = "ErrorUnsubscriberWithContext"
			}
			rows = append(rows, []string{
				code(t.String()),
				code(closer),
			})
		}
		if len(rows) > 0 {
			md.Println("## Closers")
			md.Println("")
			md.Println("List of all registered closers that are being called during graceful shutdown.")
			md.Println("")
			md.Table([]string{"Name", "Type"}, rows)
			md.Println("")
		}
	}

	{
		var rows [][]string
		s.meter.AsyncFloat64()

		var names []string
		values := map[string]nonrecording.Metric{}
		for _, value := range nonrecording.Metrics() {
			names = append(names, value.Name)
			values[value.Name] = value
		}

		gatherer, _ := prometheus.DefaultRegisterer.(*prometheus.Registry).Gather()
		for _, value := range gatherer {
			names = append(names, value.GetName())
			values[value.GetName()] = nonrecording.Metric{
				Name: value.GetName(),
				Type: value.GetType().String(),
				Help: value.GetHelp(),
			}
		}
		sort.Strings(names)
		for _, name := range names {
			value := values[name]
			rows = append(rows, []string{
				code(value.Name),
				value.Type,
				value.Help,
			})
		}
		if len(rows) > 0 {
			md.Println("## Metrics")
			md.Println("")
			md.Println("List of all registered metrics than are being exposed.")
			md.Println("")
			md.Table([]string{"Name", "Type", "Description"}, rows)
			md.Println("")
		}
	}

	fmt.Print(md.String())
}

type MD struct {
	value string
}

func (s *MD) Println(a ...any) {
	s.value += fmt.Sprintln(a...)
}

func (s *MD) Print(a ...any) {
	s.value += fmt.Sprint(a...)
}

func (s *MD) String() string {
	return s.value
}

func (s *MD) Table(headers []string, rows [][]string) {
	table, err := markdowntable.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(headers...).
		Format(rows)
	if err != nil {
		panic(err)
	}
	s.Print(table)
}

func code(v string) string {
	if v == "" {
		return ""
	}
	return "`" + v + "`"
}

func stringer(v any) string {
	if i, ok := v.(fmt.Stringer); ok {
		return i.String()
	}
	return ""
}
