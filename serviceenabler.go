package keel

import (
	"context"
	"sync"
	"time"

	"github.com/foomo/keel/interfaces"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// ServiceEnabler is a [Service] that starts and stops a wrapped service at
// runtime depending on the result of an enabled function. The enabled function
// is polled once per second; when it flips to true a fresh service is created
// with the [ServiceFn] and started, when it flips to false the service is
// closed. Failing to start or close a service dynamically is fatal.
type ServiceEnabler struct {
	l               *zap.Logger
	ctx             context.Context
	name            string
	service         Service
	serviceFn       ServiceFn
	syncEnabled     bool
	syncEnabledLock sync.RWMutex
	enabledFn       func() bool
	syncClosed      bool
	syncClosedLock  sync.RWMutex
}

// NewServiceEnabler returns a [ServiceEnabler] named name that creates its
// service with serviceFn and reads its state from enabledFn. enabledFn is
// called once immediately to determine the initial state.
func NewServiceEnabler(l *zap.Logger, name string, serviceFn ServiceFn, enabledFn func() bool) *ServiceEnabler {
	return &ServiceEnabler{
		l:           log.WithServiceName(l, name),
		name:        name,
		serviceFn:   serviceFn,
		syncEnabled: enabledFn(),
		enabledFn:   enabledFn,
	}
}

// Name returns the name of the service enabler.
func (w *ServiceEnabler) Name() string {
	return w.name
}

// Start starts watching the enabled function and starts the wrapped service if
// it is enabled. While the wrapped service is enabled, Start blocks for as long
// as the wrapped service's Start does and returns its error.
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

// Close stops watching the enabled function and closes the wrapped service if
// it is enabled. The wrapped service is closed with the context passed to
// Start rather than ctx.
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

	if i, ok := interfaces.IsCloser(w.service); ok {
		i.Close()
		return nil
	}

	if i, ok := interfaces.IsErrorCloser(w.service); ok {
		return i.Close()
	}

	if i, ok := interfaces.IsCloserWithContext(w.service); ok {
		i.Close(ctx)
		return nil
	}

	if i, ok := interfaces.IsErrorCloserWithContext(w.service); ok {
		return i.Close(ctx)
	}

	return nil
}

// watch polls enabledFn once per second until the enabler is closed and enables
// or disables the wrapped service whenever its result changes.
func (w *ServiceEnabler) watch(ctx context.Context) {
	go func() { //nolint:gosec
		for {
			time.Sleep(time.Second)

			if w.closed() {
				break
			}

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
