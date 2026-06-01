package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gofuncy"
	"github.com/foomo/keel/integration/nats/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	testingx.Tags(t, tagx.Short)
	t.Parallel()

	port := testingx.FreePort(t)

	s, err := service.NewEmbeddedServer(
		service.EmbeddedServerWithPort(port),
		service.EmbeddedServerWithHost("localhost"),
	)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("nats://localhost:%d", port), s.Server().ClientURL())

	gofuncy.Start(t.Context(), func(ctx context.Context) error {
		return s.Start(ctx)
	})

	require.Eventually(t, s.Server().Running, time.Second, 100*time.Millisecond)

	require.NoError(t, s.Close(t.Context()))

	require.Eventually(t, func() bool { return !s.Server().Running() }, time.Second, 100*time.Millisecond)
}
