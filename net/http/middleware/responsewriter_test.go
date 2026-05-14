package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foomo/keel/net/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteHeader_Idempotent(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wr := middleware.WrapResponseWriter(rec)

	wr.WriteHeader(http.StatusCreated)
	wr.WriteHeader(http.StatusInternalServerError) // should be ignored

	assert.Equal(t, http.StatusCreated, wr.StatusCode())
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestWrite_SetsImplicitHeader(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wr := middleware.WrapResponseWriter(rec)

	n, err := wr.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, http.StatusOK, wr.StatusCode())
	assert.Equal(t, 5, wr.Size())
}

func TestWrite_ThenWriteHeader_NoSuperfluous(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wr := middleware.WrapResponseWriter(rec)

	_, err := wr.Write([]byte("hello"))
	require.NoError(t, err)

	// This should be silently ignored, not cause "superfluous WriteHeader"
	wr.WriteHeader(http.StatusInternalServerError)

	assert.Equal(t, http.StatusOK, wr.StatusCode())
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUnwrap_ReturnsUnderlying(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wr := middleware.WrapResponseWriter(rec)

	assert.Equal(t, rec, wr.Unwrap())
}

type flusherRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (f *flusherRecorder) Flush() {
	f.flushed = true
}

func TestFlush_DelegatesToFlusher(t *testing.T) {
	t.Parallel()

	rec := &flusherRecorder{ResponseRecorder: httptest.NewRecorder()}
	wr := middleware.WrapResponseWriter(rec)

	wr.Flush()

	assert.True(t, rec.flushed)
}

func TestFlush_NoopWhenNotFlusher(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wr := middleware.WrapResponseWriter(rec)

	// Should not panic
	assert.NotPanics(t, func() {
		wr.Flush()
	})
}

func TestFlush_ViaTypeAssertion(t *testing.T) {
	t.Parallel()

	rec := &flusherRecorder{ResponseRecorder: httptest.NewRecorder()}
	wr := middleware.WrapResponseWriter(rec)

	f, ok := http.ResponseWriter(wr).(http.Flusher)
	require.True(t, ok, "responseWriter should implement http.Flusher")

	f.Flush()

	assert.True(t, rec.flushed)
}

func TestResponseController_Flush(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := middleware.WrapResponseWriter(w)
		rc := http.NewResponseController(wr)

		wr.WriteHeader(http.StatusOK)

		_, writeErr := wr.Write([]byte("data: hello\n\n"))
		assert.NoError(t, writeErr)

		flushErr := rc.Flush()
		assert.NoError(t, flushErr, "ResponseController.Flush should work through the wrapper")
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
