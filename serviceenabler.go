package keel

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type ServiceFunc func() Service

type ServiceEnabler struct {
	l               *zap.Logger
	ctx             context.Context
	name            string
	service         Service
	serviceFn       ServiceFunc
	syncEnabled     bool
	syncEnabledLock sync.RWMutex
	enabledFn       func() bool
	syncClosed      bool
	syncClosedLock  sync.RWMutex
}

func NewServiceEnabler(l *zap.Logger, name string, serviceFn ServiceFunc, enabledFn func() bool) *ServiceEnabler {
	return &ServiceEnabler{
		l:           log.WithServiceName(l, name),
		name:        name,
		serviceFn:   serviceFn,
		syncEnabled: enabledFn(),
		enabledFn:   enabledFn,
	}
}

func (w *ServiceEnabler) Name() string {
	return w.name
}

func (w *ServiceEnabler) Start(ctx context.Context) error {
	w.ctx = ctx
	w.watch(w.ctx) //nolint:contextcheck
	if w.enabled() {
		if err := w.enable(w.ctx); err != nil { //nolint:contextcheck
			return err
		}
	} else {
		w.l.Info("skipping disabled dynamic service")
	}
	return nil
}

func (w *ServiceEnabler) Close(ctx context.Context) error {
	l := log.WithServiceName(w.l, w.Name())
	w.setClosed(true)
	if w.enabled() {
		if err := w.disable(w.ctx); err != nil { //nolint:contextcheck
			return err
		}
	} else {
		l.Info("skipping disabled dynamic service")
	}
	return nil
}

func (w *ServiceEnabler) closed() bool {
	w.syncClosedLock.RLock()
	defer w.syncClosedLock.RUnlock()
	return w.syncClosed
}

func (w *ServiceEnabler) setClosed(v bool) {
	w.syncClosedLock.Lock()
	defer w.syncClosedLock.Unlock()
	w.syncClosed = v
}

func (w *ServiceEnabler) enabled() bool {
	w.syncEnabledLock.RLock()
	defer w.syncEnabledLock.RUnlock()
	return w.syncEnabled
}

func (w *ServiceEnabler) setEnabled(v bool) {
	w.syncEnabledLock.Lock()
	defer w.syncEnabledLock.Unlock()
	w.syncEnabled = v
}

func (w *ServiceEnabler) enable(ctx context.Context) error {
	w.setEnabled(true)
	w.service = w.serviceFn()
	w.l.Info("starting dynamic service")
	return w.service.Start(ctx)
}

func (w *ServiceEnabler) disable(ctx context.Context) error {
	w.setEnabled(false)
	w.l.Info("stopping dynamic service")
	return w.service.Close(ctx)
}

func (w *ServiceEnabler) watch(ctx context.Context) {
	go func() {
		for {
			if w.closed() {
				break
			}
			time.Sleep(time.Second)
			if value := w.enabledFn(); value != w.enabled() {
				if value {
					go func() {
						if err := w.enable(ctx); err != nil {
							w.l.Fatal("failed to dynamically start service", log.FError(err))
						}
					}()
				} else {
					if err := w.disable(context.TODO()); err != nil { //nolint:contextcheck
						w.l.Fatal("failed to dynamically close service", log.FError(err))
					}
				}
			}
		}
	}()
}
