package nats_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/gofuncy"
	"github.com/foomo/keel/integration/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	testingx.Tags(t, tagx.Short)
	t.Parallel()

	port := testingx.FreePort(t)

	s, err := nats.NewEmbeddedServer(
		nats.EmbeddedServerWithPort(port),
		nats.EmbeddedServerWithHost("localhost"),
	)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("nats://localhost:%d", port), s.Server().ClientURL())

	gofuncy.Start(t.Context(), func(ctx context.Context) error {
		return s.Start(t.Context())
	})

	testingx.WaitFor(t, time.Second, func() bool {
		return s.Server().Running()
	})

	require.NoError(t, s.Close(t.Context()))

	testingx.WaitFor(t, time.Second, func() bool {
		return !s.Server().Running()
	})
}
