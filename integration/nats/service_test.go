package nats_test

import (
	"fmt"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/keel/integration/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	testingx.Tags(t, tagx.Short)
	t.Parallel()

	port := testingx.FreePort(t)

	s, err := nats.NewService(nats.ServiceWithPort(port))
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("nats://0.0.0.0:%d", port), s.Server().ClientURL())

	err = s.Start(t.Context())
	require.NoError(t, err)

	assert.True(t, s.Server().Running())

	require.NoError(t, s.Close(t.Context()))

	assert.False(t, s.Server().Running())
}
