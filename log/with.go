package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	foomosemconv "github.com/foomo/opentelemetry-go/semconv"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
	keelsemconv "github.com/foomo/keel/semconv"
)

// With returns a child of l with fields added. A nil l selects [Logger].
func With(l *zap.Logger, fields ...zap.Field) *zap.Logger {
	if l == nil {
		l = Logger()
	}

	return l.With(fields...)
}

// WithAttributes returns a child of l with attrs added as fields, with "."
// in keys replaced by "_".
func WithAttributes(l *zap.Logger, attrs ...attribute.KeyValue) *zap.Logger {
	if l == nil {
		l = Logger()
	}

	fields := make([]zap.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = zap.Any(strings.ReplaceAll(string(attr.Key), ".", "_"), attr.Value.AsInterface())
	}

	return l.With(fields...)
}

// WithError returns a child of l with the semconv error type and exception
// message of err. A nil err adds nothing.
func WithError(l *zap.Logger, err error) *zap.Logger {
	if err == nil {
		return With(l)
	}

	return WithAttributes(l, foomosemconv.ErrorType(err), semconv.ExceptionMessage(err.Error()))
}

// WithServiceName returns a child of l with the semconv service name.
func WithServiceName(l *zap.Logger, name string) *zap.Logger {
	return With(l, Attribute(semconv.ServiceName(name)))
}

// WithTraceID returns a child of l with the trace and span ids of the span
// in ctx, or l unchanged if that span is invalid or not sampled.
func WithTraceID(l *zap.Logger, ctx context.Context) *zap.Logger {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() && spanCtx.IsSampled() {
		l = WithAttributes(l,
			keelsemconv.TraceID(spanCtx.TraceID().String()),
			keelsemconv.SpanID(spanCtx.SpanID().String()),
		)
	}

	return l
}

// WithHTTPServerName returns a child of l with name under [HTTPServerNameKey].
func WithHTTPServerName(l *zap.Logger, name string) *zap.Logger {
	return With(l, FHTTPServerName(name))
}

// WithHTTPFlavor returns a child of l with the network protocol name "HTTP"
// and the protocol version of r.
func WithHTTPFlavor(l *zap.Logger, r *http.Request) *zap.Logger {
	return With(l, Attributes(semconv.NetworkProtocolName("HTTP"), semconv.NetworkProtocolVersion(fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor)))...)
}

// WithHTTPScheme returns a child of l with the URL scheme of r, "https" if r
// was received over TLS and "http" otherwise.
func WithHTTPScheme(l *zap.Logger, r *http.Request) *zap.Logger {
	if r.TLS != nil {
		return With(l, Attribute(semconv.URLScheme("https")))
	} else {
		return With(l, Attribute(semconv.URLScheme("http")))
	}
}

// WithHTTPSessionID returns a child of l with the session id from the
// X-Session-Id header or the request context, or l unchanged if none is set.
func WithHTTPSessionID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Session-Id"); id != "" {
		return With(l, Attribute(semconv.SessionID(id)))
	} else if id, ok := keelhttpcontext.GetSessionID(r.Context()); ok && id != "" {
		return With(l, Attribute(semconv.SessionID(id)))
	} else {
		return l
	}
}

// WithHTTPRequestID returns a child of l with the request id from the
// X-Request-ID header or the request context, or l unchanged if none is set.
func WithHTTPRequestID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return WithAttributes(l, semconv.HTTPRequestHeader("x-request-id", id))
	} else if id, ok := keelhttpcontext.GetRequestID(r.Context()); ok && id != "" {
		return WithAttributes(l, semconv.HTTPRequestHeader("x-request-id", id))
	} else {
		return l
	}
}

// WithHTTPReferer returns a child of l with the referer from the X-Referer
// or Referer header, or l unchanged if none is set.
func WithHTTPReferer(l *zap.Logger, r *http.Request) *zap.Logger {
	if value := r.Header.Get("X-Referer"); value != "" {
		return WithAttributes(l, semconv.HTTPRequestHeader("referer", value))
	} else if value := r.Referer(); value != "" {
		return WithAttributes(l, semconv.HTTPRequestHeader("referer", value))
	} else {
		return l
	}
}

// WithHTTPHost returns a child of l with the host name from the
// X-Forwarded-Host header, falling back to r.Host, or the URL host for
// absolute request URLs.
func WithHTTPHost(l *zap.Logger, r *http.Request) *zap.Logger {
	if value := r.Header.Get("X-Forwarded-Host"); value != "" {
		return With(l, Attribute(semconv.HostName(value)))
	} else if !r.URL.IsAbs() {
		return With(l, Attribute(semconv.HostName(r.Host)))
	} else {
		return With(l, Attribute(semconv.HostName(r.URL.Host)))
	}
}

// WithHTTPTrackingID returns a child of l with the tracking id from the
// X-Tracking-Id header or the request context, or l unchanged if none is set.
func WithHTTPTrackingID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Tracking-Id"); id != "" {
		return With(l, Attribute(foomosemconv.TrackingID(id)))
	} else if id, ok := keelhttpcontext.GetTrackingID(r.Context()); ok && id != "" {
		return With(l, Attribute(foomosemconv.TrackingID(id)))
	} else {
		return l
	}
}

// WithHTTPClientIP returns a child of l with the client address taken from
// the first X-Forwarded-For entry, X-Real-IP or r.RemoteAddr, in that order.
func WithHTTPClientIP(l *zap.Logger, r *http.Request) *zap.Logger {
	var clientIP string

	if value := r.Header.Get("X-Forwarded-For"); value != "" {
		if i := strings.IndexAny(value, ", "); i > 0 {
			clientIP = value[:i]
		} else {
			clientIP = value
		}
	} else if value := r.Header.Get("X-Real-IP"); value != "" {
		clientIP = value
	} else if value, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		clientIP = value
	} else {
		clientIP = r.RemoteAddr
	}

	if clientIP != "" {
		return With(l, Attribute(semconv.ClientAddress(clientIP)))
	}

	return l
}

// WithHTTPRequest returns a child of l with the fields describing the
// incoming request r: host, referer, request, session and tracking ids,
// scheme, protocol, client address, trace ids, path, user agent, method and
// size.
func WithHTTPRequest(l *zap.Logger, r *http.Request) *zap.Logger {
	l = WithHTTPHost(l, r)
	l = WithHTTPReferer(l, r)
	l = WithHTTPRequestID(l, r)
	l = WithHTTPSessionID(l, r)
	l = WithHTTPTrackingID(l, r)
	l = WithHTTPScheme(l, r)
	l = WithHTTPFlavor(l, r)
	l = WithHTTPClientIP(l, r)
	l = WithTraceID(l, r.Context())

	return With(l, Attributes(
		semconv.URLPath(r.URL.Path),
		semconv.UserAgentName(r.UserAgent()),
		semconv.HTTPRequestMethodKey.String(r.Method),
		semconv.HTTPRequestSizeKey.Int64(r.ContentLength),
	)...)
}

// WithHTTPRequestOut is like [WithHTTPRequest] for an outgoing request r,
// omitting referer and client address.
func WithHTTPRequestOut(l *zap.Logger, r *http.Request) *zap.Logger {
	l = WithHTTPHost(l, r)
	l = WithHTTPRequestID(l, r)
	l = WithHTTPSessionID(l, r)
	l = WithHTTPTrackingID(l, r)
	l = WithHTTPScheme(l, r)
	l = WithHTTPFlavor(l, r)
	l = WithTraceID(l, r.Context())

	return With(l, Attributes(
		semconv.URLPath(r.URL.Path),
		semconv.UserAgentName(r.UserAgent()),
		semconv.HTTPRequestMethodKey.String(r.Method),
		semconv.HTTPRequestSizeKey.Int64(r.ContentLength),
	)...)
}
