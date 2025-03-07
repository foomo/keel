package jetstream_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/keel/net/stream/jetstream"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	natscontainer "github.com/testcontainers/testcontainers-go/modules/nats"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	testingx.Tags(t, tagx.Skip)

	l := zaptest.NewLogger(t)
	ctx := t.Context()

	natsContainer, err := natscontainer.Run(ctx, "nats:2.9.20-alpine")
	require.NoError(t, err)

	// Clean up the container
	defer func() {
		require.NoError(t, natsContainer.Terminate(ctx))
	}()

	cfg := &nats.StreamConfig{
		Replicas:   1,
		Retention:  nats.InterestPolicy,
		Duplicates: time.Second,
		MaxAge:     4 * time.Hour,
		Storage:    nats.MemoryStorage,
		Discard:    nats.DiscardOld,
		Subjects:   []string{"demo"},
	}

	// create stream
	l.Info("creating stream")
	js, err := jetstream.New(l, "test", natsContainer.MustConnectionString(ctx), jetstream.WithConfig(cfg))
	require.NoError(t, err)

	// create publisher
	l.Info("sending message #1")
	pub := js.Publisher("demo")
	_, err = pub.PublishMsg("Hello World #1")
	require.NoError(t, err)

	// create receiver
	l.Info("creating receiver")
	sub := js.Subscriber("demo")
	messages := make(chan string)
	subscription, err := sub.Subscribe(func(ctx context.Context, l *zap.Logger, msg *nats.Msg) error {
		l.Info("received message", zap.String("msg", string(msg.Data)))
		messages <- string(msg.Data)
		return nil
	},
		nats.ConsumerName("my-consumer"),
	)
	require.NoError(t, err)

	send(t, l, "Hello World #2", pub, messages)

	natsContainerName, err := natsContainer.Container.Name(ctx)
	natsContainerName = strings.TrimPrefix(natsContainerName, "/")
	require.NoError(t, err)

	{
		l.Info("pausing nats", zap.String("name", natsContainerName))
		require.NoError(t, exec.CommandContext(ctx, "docker", "pause", natsContainerName).Run())
		require.Eventually(t, func() bool {
			require.NoError(t, js.Conn().ForceReconnect())
			return !js.Conn().IsConnected()
		}, time.Minute, 100*time.Millisecond)
	}

	{
		l.Info("resuming nats", zap.String("name", natsContainerName))
		require.NoError(t, exec.CommandContext(ctx, "docker", "unpause", natsContainerName).Run())
		require.Eventually(t, func() bool {
			return js.Conn().IsConnected()
		}, time.Minute, 100*time.Millisecond)
	}

	send(t, l, "Hello World #3", pub, messages)

	// cleanup
	{
		l.Info("shutting down")
		require.NoError(t, subscription.Unsubscribe())
		js.Close()
	}
}

func send(t *testing.T, l *zap.Logger, msg string, pub *jetstream.Publisher, messages <-chan string) {
	t.Helper()

	l.Info("sending message", zap.String("msg", msg))
	_, err := pub.PublishMsg(msg)
	require.NoError(t, err)

	// wait for
	l.Info("waiting for message")
	assert.Eventually(t, func() bool {
		l.Info("received message", zap.String("msg", <-messages))
		return true
	}, time.Second, 100*time.Millisecond)
}
