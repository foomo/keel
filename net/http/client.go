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

// HTTPClientOption configures an [http.Client] created by [NewInternalHTTPClient],
// [NewExternalHTTPClient] or [NewHTTPClient]. Options are applied in order.
//
// Transport options (such as [HTTPClientWithMaxConnsPerHost]) modify the
// client's [*http.Transport] and panic if the transport has already been
// replaced by a wrapping option such as [HTTPClientWithTelemetry] or
// [HTTPClientWithRoundTripware], or by a non-[*http.Transport] set through
// [HTTPClientWithTransport]. Apply wrapping options last.
type HTTPClientOption func(*http.Client)

// ------------------------------------------------------------------
// Client options
// ------------------------------------------------------------------

// HTTPClientWithTimeout sets the client's overall request timeout
// ([http.Client.Timeout]).
func HTTPClientWithTimeout(o time.Duration) HTTPClientOption {
	return func(v *http.Client) {
		v.Timeout = o
	}
}

// HTTPClientWithJar sets the client's cookie jar.
func HTTPClientWithJar(o http.CookieJar) HTTPClientOption {
	return func(v *http.Client) {
		v.Jar = o
	}
}

// HTTPClientWithTransport replaces the client's transport. Transport options
// applied afterwards panic unless o is an [*http.Transport].
func HTTPClientWithTransport(o http.RoundTripper) HTTPClientOption {
	return func(v *http.Client) {
		v.Transport = o
	}
}

// HTTPClientWithCheckRedirect sets the client's redirect policy
// ([http.Client.CheckRedirect]).
func HTTPClientWithCheckRedirect(o func(req *http.Request, via []*http.Request) error) HTTPClientOption {
	return func(v *http.Client) {
		v.CheckRedirect = o
	}
}

// HTTPClientWithoutCrossHostRedirects stops the client following a redirect to a
// different host, and caps redirect depth at v. When either limit is hit the
// last response is returned with its body unclosed, as for
// [http.ErrUseLastResponse].
func HTTPClientWithoutCrossHostRedirects(v int) HTTPClientOption {
	return HTTPClientWithCheckRedirect(func(req *http.Request, via []*http.Request) error {
		if len(via) >= v || req.URL.Host != via[0].URL.Host {
			return http.ErrUseLastResponse
		}

		return nil
	})
}

// ------------------------------------------------------------------
// Transport options
// ------------------------------------------------------------------

// HTTPClientWithProxy sets the transport's proxy selection function.
func HTTPClientWithProxy(o func(request *http.Request) (*url.URL, error)) HTTPClientOption {
	return httpTransportOption("HTTPClientWithProxy", func(t *http.Transport) {
		t.Proxy = o
	})
}

// HTTPClientWithDialContext sets the transport's dial function for
// unencrypted TCP connections.
func HTTPClientWithDialContext(o func(ctx context.Context, network, addr string) (net.Conn, error)) HTTPClientOption {
	return httpTransportOption("HTTPClientWithDialContext", func(t *http.Transport) {
		t.DialContext = o
	})
}

// HTTPClientWithDialTLSContext sets the transport's dial function for TLS
// connections to non-proxied HTTPS requests.
func HTTPClientWithDialTLSContext(o func(ctx context.Context, network, addr string) (net.Conn, error)) HTTPClientOption {
	return httpTransportOption("HTTPClientWithDialTLSContext", func(t *http.Transport) {
		t.DialTLSContext = o
	})
}

// HTTPClientWithTLSClientConfig sets the transport's TLS configuration.
func HTTPClientWithTLSClientConfig(o *tls.Config) HTTPClientOption {
	return httpTransportOption("HTTPClientWithTLSClientConfig", func(t *http.Transport) {
		t.TLSClientConfig = o
	})
}

// HTTPClientWithTLSHandshakeTimeout sets the transport's TLS handshake
// timeout.
func HTTPClientWithTLSHandshakeTimeout(o time.Duration) HTTPClientOption {
	return httpTransportOption("HTTPClientWithTLSHandshakeTimeout", func(t *http.Transport) {
		t.TLSHandshakeTimeout = o
	})
}

// HTTPClientWithDisableKeepAlives sets whether the transport disables HTTP
// keep-alives and uses each connection for a single request only.
func HTTPClientWithDisableKeepAlives(o bool) HTTPClientOption {
	return httpTransportOption("HTTPClientWithDisableKeepAlives", func(t *http.Transport) {
		t.DisableKeepAlives = o
	})
}

// HTTPClientWithDisableCompression sets whether the transport refrains from
// requesting gzip-compressed responses.
func HTTPClientWithDisableCompression(o bool) HTTPClientOption {
	return httpTransportOption("HTTPClientWithDisableCompression", func(t *http.Transport) {
		t.DisableCompression = o
	})
}

// HTTPClientWithMaxIdleConns caps pooled idle connections across all hosts.
// Zero means no limit.
func HTTPClientWithMaxIdleConns(o int) HTTPClientOption {
	return httpTransportOption("HTTPClientWithMaxIdleConns", func(t *http.Transport) {
		t.MaxIdleConns = o
	})
}

// HTTPClientWithMaxIdleConnsPerHost caps pooled idle connections per host.
//
// The net/http default is 2, which serializes concurrent calls and causes
// constant dial churn. For in-cluster upstreams it also concentrates all load
// on two pods, because kube-proxy chooses a backend per connection.
func HTTPClientWithMaxIdleConnsPerHost(o int) HTTPClientOption {
	return httpTransportOption("HTTPClientWithMaxIdleConnsPerHost", func(t *http.Transport) {
		t.MaxIdleConnsPerHost = o
	})
}

// HTTPClientWithMaxConnsPerHost caps total (idle plus in-use) connections per
// host, providing backpressure instead of unbounded fan-out during a latency
// spike. Requests block waiting for a free connection once the cap is reached.
func HTTPClientWithMaxConnsPerHost(o int) HTTPClientOption {
	return httpTransportOption("HTTPClientWithMaxConnsPerHost", func(t *http.Transport) {
		t.MaxConnsPerHost = o
	})
}

// HTTPClientWithIdleConnTimeout sets how long an idle pooled connection is kept.
//
// Keep it below the peer's idle timeout, otherwise the client races the remote
// FIN and observes spurious EOF on the next request. Relevant peer values:
// http.Server.IdleTimeout (unbounded by default, so set it explicitly),
// AWS ALB 60s, conntrack/NAT 300s.
func HTTPClientWithIdleConnTimeout(o time.Duration) HTTPClientOption {
	return httpTransportOption("HTTPClientWithIdleConnTimeout", func(t *http.Transport) {
		t.IdleConnTimeout = o
	})
}

// HTTPClientWithResponseHeaderTimeout bounds the wait between finishing the
// request and receiving the response status line. It catches upstreams that
// accept the connection and then go quiet. It does not bound reading the body,
// so it is safe for streaming responses.
func HTTPClientWithResponseHeaderTimeout(o time.Duration) HTTPClientOption {
	return httpTransportOption("HTTPClientWithResponseHeaderTimeout", func(t *http.Transport) {
		t.ResponseHeaderTimeout = o
	})
}

// HTTPClientWithExpectContinueTimeout sets how long the transport waits for
// a server's first response headers after sending "Expect: 100-continue".
func HTTPClientWithExpectContinueTimeout(o time.Duration) HTTPClientOption {
	return httpTransportOption("HTTPClientWithExpectContinueTimeout", func(t *http.Transport) {
		t.ExpectContinueTimeout = o
	})
}

// HTTPClientWithTLSNextProto sets the transport's TLS ALPN protocol upgrade
// handlers ([http.Transport.TLSNextProto]).
func HTTPClientWithTLSNextProto(o map[string]func(authority string, c *tls.Conn) http.RoundTripper) HTTPClientOption {
	return httpTransportOption("HTTPClientWithTLSNextProto", func(t *http.Transport) {
		t.TLSNextProto = o
	})
}

// HTTPClientWithProxyConnectHeader sets the headers sent to proxies in
// CONNECT requests.
func HTTPClientWithProxyConnectHeader(o http.Header) HTTPClientOption {
	return httpTransportOption("HTTPClientWithProxyConnectHeader", func(t *http.Transport) {
		t.ProxyConnectHeader = o
	})
}

// HTTPClientWithGetProxyConnectHeader sets the function returning headers
// sent to a proxy in CONNECT requests.
func HTTPClientWithGetProxyConnectHeader(o func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error)) HTTPClientOption {
	return httpTransportOption("HTTPClientWithGetProxyConnectHeader", func(t *http.Transport) {
		t.GetProxyConnectHeader = o
	})
}

// HTTPClientWithMaxResponseHeaderBytes caps response headers from a peer we do
// not control. The net/http default is a permissive 10 MiB.
func HTTPClientWithMaxResponseHeaderBytes(o int64) HTTPClientOption {
	return httpTransportOption("HTTPClientWithMaxResponseHeaderBytes", func(t *http.Transport) {
		t.MaxResponseHeaderBytes = o
	})
}

// HTTPClientWithWriteBufferSize sets the size of the transport's write
// buffer in bytes.
func HTTPClientWithWriteBufferSize(o int) HTTPClientOption {
	return httpTransportOption("HTTPClientWithWriteBufferSize", func(t *http.Transport) {
		t.WriteBufferSize = o
	})
}

// HTTPClientWithReadBufferSize sets the size of the transport's read buffer
// in bytes.
func HTTPClientWithReadBufferSize(o int) HTTPClientOption {
	return httpTransportOption("HTTPClientWithReadBufferSize", func(t *http.Transport) {
		t.ReadBufferSize = o
	})
}

// HTTPClientWithForceAttemptHTTP2 enables HTTP/2 negotiation.
//
// The transport consults Protocols first, so this setting is ignored when
// Protocols is set, which is the case for [NewInternalHTTPClient] and
// [NewExternalHTTPClient].
//
// Deprecated: Use [HTTPClientWithProtocols] instead.
func HTTPClientWithForceAttemptHTTP2(o bool) HTTPClientOption {
	return httpTransportOption("HTTPClientWithForceAttemptHTTP2", func(t *http.Transport) {
		t.ForceAttemptHTTP2 = o
	})
}

// HTTPClientWithProtocols pins the wire protocols. Requires Go 1.24 or newer.
func HTTPClientWithProtocols(o *http.Protocols) HTTPClientOption {
	return httpTransportOption("HTTPClientWithProtocols", func(t *http.Transport) {
		t.Protocols = o
	})
}

// ------------------------------------------------------------------
// Wrapping options — apply these last
// ------------------------------------------------------------------

// HTTPClientWithRoundTripware wraps the transport with the given
// [roundtripware.RoundTripware], using l for logging. Apply after all
// transport-tuning options.
func HTTPClientWithRoundTripware(l *zap.Logger, roundTripware ...roundtripware.RoundTripware) HTTPClientOption {
	return func(v *http.Client) {
		v.Transport = roundtripware.NewRoundTripper(l, v.Transport, roundTripware...)
	}
}

// HTTPClientWithTelemetry wraps the transport with OTel instrumentation. Apply
// after all transport-tuning options.
//
// If a retrying round tripper is also in play, put telemetry outermost so it
// reports one span per logical call rather than one per attempt.
//
// Spans are named "HTTP GET" by default. Pass [otelhttp.WithSpanNameFormatter]
// returning "{method} {url.template}" (e.g. "GET /prices/{sku}"); never use
// the URL path, as it makes span names unbounded.
func HTTPClientWithTelemetry(opts ...otelhttp.Option) HTTPClientOption {
	return func(v *http.Client) {
		v.Transport = otelhttp.NewTransport(v.Transport, opts...)
	}
}

// ------------------------------------------------------------------
// Dialer and transport
// ------------------------------------------------------------------

// DefaultHTTPTransportDialer returns the legacy dialer used by
// [DefaultHTTPTransport].
//
// It uses a 45s dial timeout and keep-alive period. A 45s dial timeout is far
// too slack for in-cluster traffic, where a slow dial means the endpoint is
// gone rather than far away, and the loose keep-alive takes minutes to notice a
// pod that disappeared without a FIN.
//
// Deprecated: Use [NewHTTPTransportDialer] instead.
func DefaultHTTPTransportDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   45 * time.Second,
		KeepAlive: 45 * time.Second,
	}
}

// NewHTTPTransportDialer returns the dialer used by [NewHTTPTransport].
//
// internal selects tuning for pod-to-pod calls (2s dial timeout); false targets
// the public internet (5s dial timeout). Both enable aggressive TCP keep-alive
// probing (15s idle, 5s interval, 3 probes).
func NewHTTPTransportDialer(internal bool) *net.Dialer {
	return &net.Dialer{
		// In-cluster the peer is one network hop away, so a slow dial means the
		// endpoint no longer exists rather than that the network is far.
		Timeout: pick(internal, 2*time.Second, 5*time.Second),

		// TCP keep-alive probes, unrelated to HTTP keep-alive. The Go default
		// can take minutes to notice a pod that disappeared without sending a
		// FIN — deleted, OOM-killed, node lost. Aggressive probing turns a
		// black hole into a prompt connection error a retry can act on.
		//
		// Requires Go 1.23+. On older versions use KeepAlive: 15 * time.Second.
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   true,
			Idle:     15 * time.Second,
			Interval: 5 * time.Second,
			Count:    3,
		},
	}
}

// DefaultHTTPTransport returns the legacy transport used by [NewHTTPClient].
//
// This transport disables connection pooling and leaves MaxIdleConnsPerHost
// (defaulting to 2), MaxConnsPerHost, IdleConnTimeout and
// ResponseHeaderTimeout unset.
//
// Deprecated: Use [NewHTTPTransport] instead.
func DefaultHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           DefaultHTTPTransportDialer().DialContext,
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	}
}

// NewHTTPTransport returns a tuned transport. Each call returns a distinct
// [*http.Transport] and therefore a distinct connection pool.
//
// internal selects tuning for another Deployment in the same cluster, reached
// through a ClusterIP Service with no sidecar or gateway in the path; false
// targets an API on the public internet.
//
// The two differ because the network in between differs. kube-proxy is an L4
// balancer that picks a backend pod once per TCP connection, whereas a public
// API sits behind an L7 load balancer that routes per request. That single fact
// drives the protocol choice, the pool width and the timeouts.
//
//	                          internal   external
//	dial timeout              2s         5s
//	client timeout (caller)   10s        30s
//	MaxIdleConns              256        64
//	MaxIdleConnsPerHost       64         16
//	MaxConnsPerHost           128        32
//	IdleConnTimeout           30s        60s
//	ResponseHeaderTimeout     5s         10s
//	MaxResponseHeaderBytes    1 MiB      64 KiB
//	HTTP/2                    no         yes
//	compression               off        on
func NewHTTPTransport(internal bool) *http.Transport {
	// Requires Go 1.24+. On older versions drop this and set
	// ForceAttemptHTTP2: !internal below.
	var proto http.Protocols
	proto.SetHTTP1(true)
	// No HTTP/2 for in-cluster calls. kube-proxy is an L4 balancer that pins a
	// TCP connection to a single pod, so multiplexing every request onto one
	// connection sends all traffic to one replica and scaling the target
	// Deployment achieves nothing. HTTP/1.1 with a wide pool spreads load,
	// since each connection gets its own backend.
	proto.SetHTTP2(!internal)

	return &http.Transport{
		// Honour HTTP_PROXY/HTTPS_PROXY/NO_PROXY so an egress proxy can be
		// injected via environment. NO_PROXY must cover .svc.cluster.local, or
		// in-cluster traffic is proxied too.
		Proxy:       http.ProxyFromEnvironment,
		DialContext: NewHTTPTransportDialer(internal).DialContext,

		// Connection pooling stays on. Disabling it costs a handshake per
		// request and leaves a socket in TIME_WAIT for ~60s, which exhausts the
		// ephemeral port range at a few hundred rps.
		DisableKeepAlives: false,

		MaxIdleConns:        pick(internal, 256, 64),
		MaxIdleConnsPerHost: pick(internal, 64, 16),
		MaxConnsPerHost:     pick(internal, 128, 32),

		// Below any plausible peer idle timeout, so the client is always the
		// side that closes and never races an incoming FIN.
		IdleConnTimeout: pick(internal, 30*time.Second, 60*time.Second),

		ResponseHeaderTimeout: pick(internal, 5*time.Second, 10*time.Second),
		ExpectContinueTimeout: 1 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,

		MaxResponseHeaderBytes: pick(internal, int64(1<<20), int64(64<<10)),

		// gzip on both ends for traffic that never leaves the node's network is
		// wasted CPU; over the internet it is a clear win.
		DisableCompression: internal,

		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		Protocols:       &proto,
	}
}

// ------------------------------------------------------------------
// Constructors
// ------------------------------------------------------------------

// NewInternalHTTPClient returns a client tuned for pod-to-pod calls through a
// ClusterIP Service with no mesh or gateway in the path, using
// [NewHTTPTransport] with internal set to true and a 10s timeout. Options are
// applied afterwards, so [HTTPClientWithTimeout] overrides the timeout.
//
// Create one client per logical upstream at process start and reuse it for the
// lifetime of the process: the Transport is the connection pool, so a client
// per request defeats pooling and leaks file descriptors.
func NewInternalHTTPClient(opts ...HTTPClientOption) *http.Client {
	return newHTTPClient(true, opts...)
}

// NewExternalHTTPClient returns a client tuned for third-party HTTPS APIs,
// using [NewHTTPTransport] with internal set to false and a 30s timeout.
// The same reuse rule applies as for [NewInternalHTTPClient].
func NewExternalHTTPClient(opts ...HTTPClientOption) *http.Client {
	return newHTTPClient(false, opts...)
}

// NewHTTPClient returns a client using the legacy transport from
// [DefaultHTTPTransport] and a two minute timeout. Compared to the tuned
// constructors it differs as follows:
//
//	setting                 legacy        NewExternalHTTPClient
//	DisableKeepAlives       true          false
//	MaxIdleConnsPerHost     2 (default)   16
//	MaxConnsPerHost         unlimited     32
//	IdleConnTimeout         unlimited     60s
//	ResponseHeaderTimeout   unset         10s
//	dial timeout            45s           5s
//	Timeout                 2m            30s
//
// The worst of these is DisableKeepAlives: every request pays a fresh
// handshake and leaves a socket in TIME_WAIT for ~60s, so a few hundred
// requests per second exhausts the ephemeral port range. The symptom is
// intermittent dial failures that look like a network fault.
//
// Migrating is not purely mechanical. Callers relying on the two minute timeout
// for slow endpoints must also raise ResponseHeaderTimeout, which the tuned
// defaults bound at 10s — otherwise a request that takes 45s to produce its
// first byte fails even with a generous overall timeout.
//
// Deprecated: Use [NewInternalHTTPClient] or [NewExternalHTTPClient] instead.
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

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func newHTTPClient(internal bool, opts ...HTTPClientOption) *http.Client {
	inst := &http.Client{
		Transport: NewHTTPTransport(internal),

		// Backstop only. The real deadline should come from the request
		// context, so an inbound deadline propagates to outbound calls.
		// Options are applied afterwards, so HTTPClientWithTimeout overrides it.
		Timeout: pick(internal, 10*time.Second, 30*time.Second),
	}
	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

// httpTransportOption applies f to the client's *http.Transport.
//
// Wrapping options such as HTTPClientWithTelemetry and
// HTTPClientWithRoundTripware replace Transport with a decorator, after which
// the underlying transport can no longer be reached by a type assertion. The
// returned option panics in that case rather than silently ignoring the
// setting, so a misordered option fails at construction time.
func httpTransportOption(name string, f func(*http.Transport)) HTTPClientOption {
	return func(v *http.Client) {
		t, ok := v.Transport.(*http.Transport)
		if !ok {
			panic("keel/net/http: " + name + " must be applied before wrapping options such as HTTPClientWithTelemetry or HTTPClientWithRoundTripware")
		}

		f(t)
	}
}

// pick returns a if cond is true and b otherwise.
func pick[T any](cond bool, a, b T) T {
	if cond {
		return a
	}

	return b
}
