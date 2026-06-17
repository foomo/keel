package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foomo/keel/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushToGateway(t *testing.T) {
	t.Parallel()

	var (
		gotPath   string
		gotMethod string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method

		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	err := telemetry.PushToGateway(context.Background(), srv.URL, "my-job")
	require.NoError(t, err)

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/metrics/job/my-job", gotPath)
}
