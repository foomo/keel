package keel

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type ServiceFunc func() Service

type ServiceEnabler struct {
	l         *zap.Logger
	ctx       context.Context
	name      string
	service   Service
	serviceFn ServiceFunc
	enabled   bool
	enabledFn func() bool
	closed    bool
}

func NewServiceEnabler(l *zap.Logger, name string, serviceFn ServiceFunc, enabledFn func() bool) *ServiceEnabler {
	return &ServiceEnabler{
		l:         log.WithServiceName(l, name),
		name:      name,
		serviceFn: serviceFn,
		enabled:   enabledFn(),
		enabledFn: enabledFn,
	}
}

func (w *ServiceEnabler) Name() string {
	return w.name
}

func (w *ServiceEnabler) enable(ctx context.Context) error {
	w.enabled = true
	w.service = w.serviceFn()
	w.l.Info("starting dynamic service")
	return w.service.Start(ctx)
}

func (w *ServiceEnabler) disable(ctx context.Context) error {
	w.enabled = false
	w.l.Info("stopping dynamic service")
	return w.service.Close(ctx)
}

func (w *ServiceEnabler) watch() {
	go func() {
		for {
			if w.closed {
				break
			}
			time.Sleep(time.Second)
			if value := w.enabledFn(); value != w.enabled {
				if value {
					go func() {
						if err := w.enable(w.ctx); err != nil {
							w.l.Fatal("failed to dynamically start service", log.FError(err))
						}
					}()
				} else {
					if err := w.disable(context.TODO()); err != nil {
						w.l.Fatal("failed to dynamically close service", log.FError(err))
					}
				}
			}
		}
	}()
}

func (w *ServiceEnabler) Start(ctx context.Context) error {
	w.watch()
	w.ctx = ctx
	if w.enabled {
		if err := w.enable(w.ctx); err != nil {
			return err
		}
	} else {
		w.l.Info("skipping disabled dynamic service")
	}
	return nil
}

func (w *ServiceEnabler) Close(ctx context.Context) error {
	l := log.WithServiceName(w.l, w.Name())
	w.closed = true
	if w.enabled {
		if err := w.disable(w.ctx); err != nil {
			return err
		}
	} else {
		l.Info("skipping disabled dynamic service")
	}
	return nil
}
