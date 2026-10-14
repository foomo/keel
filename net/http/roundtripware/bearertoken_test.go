package roundtripware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/foomo/keel/net/http/roundtripware"
)

func TestBearerToken(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name, token, header, existing, want string
		opts                                []roundtripware.BearerTokenOption
	}{
		{name: "default", token: "secret", header: "Authorization", want: "Bearer secret"},
		{name: "empty token", header: "Authorization"},
		{name: "existing header", token: "secret", header: "Authorization", existing: "Basic authored", want: "Basic authored"},
		{name: "custom header", token: "secret", header: "X-Auth", want: "Bearer secret", opts: []roundtripware.BearerTokenOption{roundtripware.BearerTokenWithHeader("X-Auth")}},
		{name: "custom scheme", token: "secret", header: "Authorization", want: "Token secret", opts: []roundtripware.BearerTokenOption{roundtripware.BearerTokenWithScheme("Token ")}},
		{name: "raw token", token: "secret", header: "X-Api-Key", want: "secret", opts: []roundtripware.BearerTokenOption{roundtripware.BearerTokenWithHeader("X-Api-Key"), roundtripware.BearerTokenWithScheme("")}},
		{name: "nil option", token: "secret", header: "Authorization", want: "Bearer secret", opts: []roundtripware.BearerTokenOption{nil}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.invalid", nil)
			if tt.existing != "" {
				r.Header.Set(tt.header, tt.existing)
			}

			called := false
			handler := roundtripware.BearerToken(tt.token, tt.opts...)(zap.NewNop(), func(sent *http.Request) (*http.Response, error) {
				called = true

				assert.Equal(t, tt.want, sent.Header.Get(tt.header))

				if tt.header != "Authorization" {
					assert.Empty(t, sent.Header.Get("Authorization"))
				}

				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
			})
			res, err := handler(r)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.True(t, called)
		})
	}
}

func TestBearerTokenPropagatesTransportError(t *testing.T) {
	t.Parallel()

	want := errors.New("transport failed")
	handler := roundtripware.BearerToken("secret")(zap.NewNop(), func(*http.Request) (*http.Response, error) {
		return nil, want
	})

	res, err := handler(httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.invalid", nil))
	if res != nil {
		require.NoError(t, res.Body.Close())
	}

	assert.Nil(t, res)
	require.ErrorIs(t, err, want)
}

func TestBearerTokenDoesNotRecordCredentials(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))

	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	ctx, span := provider.Tracer("test").Start(t.Context(), "request")
	core, logs := observer.New(zap.DebugLevel)
	handler := roundtripware.BearerToken("secret-token")(zap.New(core), func(*http.Request) (*http.Response, error) {
		return &http.Response{Body: http.NoBody}, nil
	})
	res, err := handler(httptest.NewRequestWithContext(ctx, http.MethodGet, "http://example.invalid", nil))
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	span.End()

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Len(t, spans[0].Events(), 1)
	assert.Equal(t, "BearerToken", spans[0].Events()[0].Name)
	data, err := json.Marshal(struct {
		Attributes any `json:"attributes"`
		Events     any `json:"events"`
	}{spans[0].Attributes(), spans[0].Events()})
	require.NoError(t, err)
	assert.NotContains(t, string(data), "secret-token")
	assert.Empty(t, logs.All())
}
