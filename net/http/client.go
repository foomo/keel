package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/foomo/keel/net/http/roundtripware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type HTTPClientOption func(*http.Client)

func HTTPClientWithTimeout(o time.Duration) HTTPClientOption {
	return func(v *http.Client) {
		v.Timeout = o
	}
}

func HTTPClientWithJar(o http.CookieJar) HTTPClientOption {
	return func(v *http.Client) {
		v.Jar = o
	}
}

func HTTPClientWithTransport(o http.RoundTripper) HTTPClientOption {
	return func(v *http.Client) {
		// TODO warn in case of overriding other options
		v.Transport = o
	}
}

func HTTPClientWithCheckRedirect(o func(req *http.Request, via []*http.Request) error) HTTPClientOption {
	return func(v *http.Client) {
		v.CheckRedirect = o
	}
}

func HTTPClientWithProxy(o func(request *http.Request) (*url.URL, error)) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.Proxy = o
			v.Transport = t
		}
	}
}

func HTTPClientWithDialContext(o func(ctx context.Context, network, addr string) (net.Conn, error)) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.DialContext = o
			v.Transport = t
		}
	}
}

func HTTPClientWithDialTLSContext(o func(ctx context.Context, network, addr string) (net.Conn, error)) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.DialTLSContext = o
			v.Transport = t
		}
	}
}

func HTTPClientWithTLSClientConfig(o *tls.Config) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.TLSClientConfig = o
			v.Transport = t
		}
	}
}

func HTTPClientWithTLSHandshakeTimeout(o time.Duration) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.TLSHandshakeTimeout = o
			v.Transport = t
		}
	}
}

func HTTPClientWithDisableKeepAlives(o bool) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.DisableKeepAlives = o
			v.Transport = t
		}
	}
}

func HTTPClientWithDisableCompression(o bool) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.DisableCompression = o
			v.Transport = t
		}
	}
}

func HTTPClientWithMaxIdleConns(o int) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.MaxIdleConns = o
			v.Transport = t
		}
	}
}

func HTTPClientWithMaxIdleConnsPerHost(o int) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.MaxIdleConnsPerHost = o
			v.Transport = t
		}
	}
}

func HTTPClientWithMaxConnsPerHost(o int) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.MaxConnsPerHost = o
			v.Transport = t
		}
	}
}

func HTTPClientWithIdleConnTimeout(o time.Duration) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.IdleConnTimeout = o
			v.Transport = t
		}
	}
}

func HTTPClientWithResponseHeaderTimeout(o time.Duration) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.ResponseHeaderTimeout = o
			v.Transport = t
		}
	}
}

func HTTPClientWithExpectContinueTimeout(o time.Duration) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.ExpectContinueTimeout = o
			v.Transport = t
		}
	}
}

func HTTPClientWithTLSNextProto(o map[string]func(authority string, c *tls.Conn) http.RoundTripper) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.TLSNextProto = o
			v.Transport = t
		}
	}
}

func HTTPClientWithProxyConnectHeader(o http.Header) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.ProxyConnectHeader = o
			v.Transport = t
		}
	}
}

func HTTPClientWithGetProxyConnectHeader(o func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error)) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.GetProxyConnectHeader = o
			v.Transport = t
		}
	}
}

func HTTPClientWithMaxResponseHeaderBytes(o int64) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.MaxResponseHeaderBytes = o
			v.Transport = t
		}
	}
}

func HTTPClientWithWriteBufferSize(o int) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.WriteBufferSize = o
			v.Transport = t
		}
	}
}

func HTTPClientWithReadBufferSize(o int) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.ReadBufferSize = o
			v.Transport = t
		}
	}
}

func HTTPClientWithForceAttemptHTTP2(o bool) HTTPClientOption {
	return func(v *http.Client) {
		if t, ok := v.Transport.(*http.Transport); ok {
			t.ForceAttemptHTTP2 = o
			v.Transport = t
		}
	}
}

func HTTPClientWithRoundTripware(l *zap.Logger, roundTripware ...roundtripware.RoundTripware) HTTPClientOption {
	return func(v *http.Client) {
		v.Transport = roundtripware.NewRoundTripper(l, v.Transport, roundTripware...)
	}
}

func HTTPClientWithTelemetry(opts ...otelhttp.Option) HTTPClientOption {
	return func(v *http.Client) {
		v.Transport = otelhttp.NewTransport(v.Transport, opts...)
	}
}

func DefaultHTTPTransportDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   45 * time.Second,
		KeepAlive: 45 * time.Second,
	}
}

func DefaultHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           DefaultHTTPTransportDialer().DialContext,
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	}
}

func NewHTTPClient(opts ...HTTPClientOption) *http.Client {
	inst := &http.Client{
		Transport: DefaultHTTPTransport(),
		Timeout:   2 * time.Minute,
	}
	for _, opt := range opts {
		opt(inst)
	}

	return inst
}
