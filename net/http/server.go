package http

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewServer creates and configures an HTTP server for in-cluster traffic
// listening on addr. handler is wrapped with middlewares using [Compose], and
// net/http errors are logged through l at error level.
//
// The server speaks HTTP/1.1 only and sets conservative timeouts:
// ReadHeaderTimeout 5s, ReadTimeout 30s, WriteTimeout 60s, IdleTimeout 620s
// and MaxHeaderBytes 1 MiB. Streaming or large-upload handlers must clear the
// read or write deadline via [http.NewResponseController].
func NewServer(l *zap.Logger, name, addr string, handler http.Handler, middlewares ...Middleware) *http.Server {
	// net/http's own errors belong at error level
	errLog, err := zap.NewStdLogAt(l.Named("http.server"), zapcore.ErrorLevel)
	if err != nil {
		errLog = zap.NewStdLog(l)
	}

	// Explicit HTTP/1.1. Clients reaching this server through a ClusterIP
	// Service are pinned to one pod per TCP connection by kube-proxy, so h2
	// multiplexing would concentrate their traffic on a single replica; the
	// client side is configured to match. Add proto.SetUnencryptedHTTP2(true)
	// if a gRPC-gateway or an Envoy probe needs h2c. Requires Go 1.24+.
	var proto http.Protocols
	proto.SetHTTP1(true)

	return &http.Server{
		Addr:     addr,
		Handler:  Compose(l, name, handler, middlewares...),
		ErrorLog: errLog,

		// Slowloris guard: bounds only the request line and headers, so it can
		// be strict without limiting body size or streaming.
		ReadHeaderTimeout: 5 * time.Second,

		// Absolute deadline for reading the whole body — it does not reset on
		// progress, so it effectively caps upload size by wall-clock time.
		// Handlers accepting large uploads should clear it via
		// http.NewResponseController and bound the body with
		// http.MaxBytesReader instead.
		ReadTimeout: 30 * time.Second,

		// Absolute deadline for writing the response, likewise not reset on
		// progress. Streaming handlers (SSE, watch, exports) must clear it with
		// http.NewResponseController(w).SetWriteDeadline(time.Time{}).
		WriteTimeout: 60 * time.Second,

		// Must exceed the idle timeout of whatever is on the other side, so that
		// side is always the one closing an idle connection. Otherwise both
		// expire together and the peer is handed a socket this server is
		// closing: intermittent EOF in-cluster, 502s behind a load balancer.
		//
		// 620s is Google's recommended backend value and clears every managed
		// L7 we run behind: GCLB 600s (fixed), Azure App Gateway 240s public /
		// 300s private, ingress-nginx 60s, ALB 60s. In-cluster clients close far
		// sooner (IdleConnTimeout 30s internal, 60s external), so they remain
		// the closing side too. An Envoy sidecar defaults to 1h and is the one
		// peer this cannot outlast — it reconnects per request, so its worst
		// case is a retry rather than a 502.
		//
		// The cost is idle connections living ~10 minutes: bounded here, since
		// the peer is always an ingress pool or a known set of pods. Lower it if
		// a pod is ever exposed by Service type=LoadBalancer or NodePort, where
		// the far end is unbounded and untrusted.
		IdleTimeout: 620 * time.Second,

		MaxHeaderBytes: 1 << 20,
		Protocols:      &proto,
	}
}
