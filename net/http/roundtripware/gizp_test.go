package roundtripware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	stdhttp "github.com/foomo/gostandards/http"
	"github.com/foomo/keel/net/http/roundtripware"
)

const (
	gzipPayload1023 = "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. \n\nDuis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisisd"
	gzipPayload1024 = "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. \n\nDuis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisisda"
)

func TestGZip(t *testing.T) {
	// create logger
	l := zaptest.NewLogger(t)

	var payload string
	var compressed bool

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate request header
		assert.Equal(t, stdhttp.EncodingGzip.String(), stdhttp.HeaderAcceptEncoding.Get(r.Header))
		if compressed {
			assert.Equal(t, stdhttp.EncodingGzip.String(), stdhttp.HeaderContentEncoding.Get(r.Header))
		} else {
			assert.Empty(t, stdhttp.HeaderContentEncoding.Get(r.Header))
		}

		// validate request body
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		if compressed {
			assert.NotEqual(t, payload, string(body))
			decomppressed, err := gunzipString(body)
			assert.NoError(t, err)
			assert.Equal(t, payload, string(decomppressed))
		} else {
			assert.Equal(t, payload, string(body))
		}

		// send response
		if compressed {
			w.Header().Set(stdhttp.HeaderContentEncoding.String(), r.Header.Get(stdhttp.HeaderContentEncoding.String()))
		}
		_, _ = w.Write(body)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.GZip(),
		),
	)

	t.Run("<1024", func(t *testing.T) {
		payload = gzipPayload1023
		compressed = false

		// create request
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, svr.URL, strings.NewReader(payload))
		req.Header.Set(stdhttp.HeaderAcceptEncoding.String(), stdhttp.EncodingGzip.String())
		require.NoError(t, err)

		// do request
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// validate repsone header
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Empty(t, resp.Header.Get(stdhttp.HeaderContentEncoding.String()))
		// validate repsone body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, payload, string(body))
	})

	t.Run(">=1024", func(t *testing.T) {
		payload = gzipPayload1024
		compressed = true

		// create request
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, svr.URL, strings.NewReader(payload))
		req.Header.Set(stdhttp.HeaderAcceptEncoding.String(), stdhttp.EncodingGzip.String())
		require.NoError(t, err)

		// do request
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// validate repsone header
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, stdhttp.EncodingGzip.String(), resp.Header.Get(stdhttp.HeaderContentEncoding.String()))
		// validate repsone body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotEqual(t, payload, string(body))

		decompressedBody, err := gunzipString(body)
		require.NoError(t, err)
		assert.Equal(t, payload, string(decompressedBody))
	})
}

func gunzipString(body []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
