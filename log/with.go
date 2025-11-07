package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	keelsemconv "github.com/foomo/keel/semconv"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

func With(l *zap.Logger, fields ...zap.Field) *zap.Logger {
	if l == nil {
		l = Logger()
	}

	return l.With(fields...)
}

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

func WithError(l *zap.Logger, err error) *zap.Logger {
	return With(l, FErrorType(err), FError(err))
}

func WithServiceName(l *zap.Logger, name string) *zap.Logger {
	return With(l, Attribute(semconv.ServiceName(name)))
}

func WithTraceID(l *zap.Logger, ctx context.Context) *zap.Logger {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() && spanCtx.IsSampled() {
		l = With(l, FTraceID(spanCtx.TraceID().String()), FSpanID(spanCtx.SpanID().String()))
	}

	return l
}

func WithHTTPServerName(l *zap.Logger, name string) *zap.Logger {
	return With(l, FHTTPServerName(name))
}

func WithHTTPFlavor(l *zap.Logger, r *http.Request) *zap.Logger {
	return With(l, Attributes(semconv.NetworkProtocolName("HTTP"), semconv.NetworkProtocolVersion(fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor)))...)
}

func WithHTTPScheme(l *zap.Logger, r *http.Request) *zap.Logger {
	if r.TLS != nil {
		return With(l, Attribute(semconv.URLScheme("https")))
	} else {
		return With(l, Attribute(semconv.URLScheme("http")))
	}
}

func WithHTTPSessionID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Session-Id"); id != "" {
		return With(l, Attribute(semconv.SessionID(id)))
	} else if id, ok := keelhttpcontext.GetSessionID(r.Context()); ok && id != "" {
		return With(l, Attribute(semconv.SessionID(id)))
	} else {
		return l
	}
}

func WithHTTPRequestID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return With(l, Attribute(keelsemconv.HTTPXRequestID(id)))
	} else if id, ok := keelhttpcontext.GetRequestID(r.Context()); ok && id != "" {
		return With(l, Attribute(keelsemconv.HTTPXRequestID(id)))
	} else {
		return l
	}
}

func WithHTTPReferer(l *zap.Logger, r *http.Request) *zap.Logger {
	if value := r.Header.Get("X-Referer"); value != "" {
		return With(l, Attribute(keelsemconv.HTTPXRequestReferer(value)))
	} else if value := r.Referer(); value != "" {
		return With(l, Attribute(keelsemconv.HTTPXRequestReferer(value)))
	} else {
		return l
	}
}

func WithHTTPHost(l *zap.Logger, r *http.Request) *zap.Logger {
	if value := r.Header.Get("X-Forwarded-Host"); value != "" {
		return With(l, Attribute(semconv.HostName(value)))
	} else if !r.URL.IsAbs() {
		return With(l, Attribute(semconv.HostName(r.Host)))
	} else {
		return With(l, Attribute(semconv.HostName(r.URL.Host)))
	}
}

func WithHTTPTrackingID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Tracking-Id"); id != "" {
		return With(l, Attribute(keelsemconv.TrackingID(id)))
	} else if id, ok := keelhttpcontext.GetTrackingID(r.Context()); ok && id != "" {
		return With(l, Attribute(keelsemconv.TrackingID(id)))
	} else {
		return l
	}
}

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
