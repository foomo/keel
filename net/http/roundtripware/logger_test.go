package roundtripware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/roundtripware"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		loggerOpts          []roundtripware.LoggerOption
		expectedLogLevel    zapcore.Level
		expectedLogMessage  string
		serverStatusCode    int
		injectError         bool
		expectResponseError bool
	}{
		{
			name:                "default logger",
			loggerOpts:          nil,
			expectedLogLevel:    zapcore.InfoLevel,
			expectedLogMessage:  "sent request",
			serverStatusCode:    http.StatusOK,
			injectError:         false,
			expectResponseError: false,
		},
		{
			name: "custom message",
			loggerOpts: []roundtripware.LoggerOption{
				roundtripware.LoggerWithMessage("my message"),
			},
			expectedLogLevel:    zapcore.InfoLevel,
			expectedLogMessage:  "my message",
			serverStatusCode:    http.StatusOK,
			injectError:         false,
			expectResponseError: false,
		},
		{
			name: "custom error message",
			loggerOpts: []roundtripware.LoggerOption{
				roundtripware.LoggerWithErrorMessage("my error message"),
			},
			expectedLogLevel:    zapcore.ErrorLevel,
			expectedLogMessage:  "my error message",
			serverStatusCode:    http.StatusOK,
			injectError:         true,
			expectResponseError: true,
		},
		{
			name: "min warn code",
			loggerOpts: []roundtripware.LoggerOption{
				roundtripware.LoggerWithMinWarnCode(200),
			},
			expectedLogLevel:    zapcore.WarnLevel,
			expectedLogMessage:  "sent request",
			serverStatusCode:    http.StatusOK,
			injectError:         false,
			expectResponseError: false,
		},
		{
			name: "min error code",
			loggerOpts: []roundtripware.LoggerOption{
				roundtripware.LoggerWithMinErrorCode(200),
			},
			expectedLogLevel:    zapcore.ErrorLevel,
			expectedLogMessage:  "sent request",
			serverStatusCode:    http.StatusOK,
			injectError:         false,
			expectResponseError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// create logger & validate output
			l := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(entry zapcore.Entry) error {
				assert.Equal(t, tt.expectedLogLevel, entry.Level)
				assert.Equal(t, tt.expectedLogMessage, entry.Message)

				return nil
			})))

			// create http server with handler
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatusCode)
			}))
			defer svr.Close()

			// build middleware chain
			middlewares := []roundtripware.RoundTripware{}
			if tt.injectError {
				middlewares = append(middlewares, func(l *zap.Logger, next roundtripware.Handler) roundtripware.Handler {
					return func(r *http.Request) (*http.Response, error) {
						return nil, errors.New("something went wrong")
					}
				})
			}

			middlewares = append(middlewares, roundtripware.Logger(tt.loggerOpts...))

			// create http client
			client := keelhttp.NewHTTPClient(
				keelhttp.HTTPClientWithRoundTripware(l, middlewares...),
			)

			// create request
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL, nil)
			require.NoError(t, err)

			// do request
			resp, err := client.Do(req)

			// validate
			if tt.expectResponseError {
				require.Nil(t, resp)
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.NotNil(t, resp)
				defer resp.Body.Close()
			}

			if resp != nil {
				defer resp.Body.Close()
			}
		})
	}
}
