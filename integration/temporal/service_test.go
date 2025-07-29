package keeltemporal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

func TestNewToggleableService(t *testing.T) {
	t.Run("NewService_StartCloseStart", func(t *testing.T) {
		l := zap.NewNop()
		client, err := client.NewLazyClient(client.Options{})
		require.NoError(t, err)

		w := worker.New(client, "test_queue", worker.Options{})
		s := NewService(l, "test", w)

		// start worker
		s.Start(t.Context())

		// close worker
		s.Close(t.Context())

		// starting worker again panics
		defer func() {
			if e := recover(); e != nil {
				if errString, ok := e.(string); ok {
					assert.Equal(t, "attempted to start a worker that has been stopped before", errString)
				} else {
					t.Fatal(e)
				}
			} else {
				t.Fatal("expected a panic")
			}
		}()

		s.Start(t.Context())
	})

	t.Run("NewToggleableService_StartCloseStart", func(t *testing.T) {
		l := zap.NewNop()
		client, err := client.NewLazyClient(client.Options{})
		require.NoError(t, err)

		newWorker := func() worker.Worker {
			return worker.New(client, "test_queue", worker.Options{})
		}
		s := NewToggleableService(l, "test", newWorker)

		// start worker
		s.Start(t.Context())

		// close worker
		s.Close(t.Context())

		// starting worker again should NOT panic
		defer func() {
			e := recover()
			assert.Nil(t, e)
		}()

		s.Start(t.Context())
	})

	t.Run("NewToggleableService_CloseBeforeStart", func(t *testing.T) {
		l := zap.NewNop()
		client, err := client.NewLazyClient(client.Options{})
		require.NoError(t, err)

		newWorker := func() worker.Worker {
			return worker.New(client, "test_queue", worker.Options{})
		}
		s := NewToggleableService(l, "test", newWorker)

		// closing worker before again should NOT panic
		defer func() {
			e := recover()
			assert.Nil(t, e)
		}()
		s.Close(t.Context())
	})

	t.Run("NewToggleableService_MultipleCallsToStart", func(t *testing.T) {
		l := zap.NewNop()
		client, err := client.NewLazyClient(client.Options{})
		require.NoError(t, err)

		newWorker := func() worker.Worker {
			return worker.New(client, "test_queue", worker.Options{})
		}
		s := NewToggleableService(l, "test", newWorker)

		// start worker
		s.Start(t.Context())

		// attempt to start worker again
		err = s.Start(t.Context())
		require.ErrorContains(t, err, "can not start new worker while another one is runnin")
	})
}
