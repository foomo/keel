package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func With(l *zap.Logger, fields ...zap.Field) *zap.Logger {
	if l == nil {
		l = Logger()
	}
	return l.With(fields...)
}

func WithError(l *zap.Logger, err error) *zap.Logger {
	return With(l, FError(err))
}

func WithHTTPServerName(l *zap.Logger, name string) *zap.Logger {
	return With(l, FHTTPServerName(name))
}

func WithServiceName(l *zap.Logger, name string) *zap.Logger {
	return With(l, FServiceName(name))
}

func WithTraceID(l *zap.Logger, ctx context.Context) *zap.Logger {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		l = With(l, FTraceID(spanCtx.TraceID().String()))
	}
	return l
}

func WithHTTPRequest(l *zap.Logger, r *http.Request) *zap.Logger {
	fields := []zap.Field{
		FHTTPRequestContentLength(r.ContentLength),
		FHTTPMethod(r.Method),
		FHTTPUserAgent(r.UserAgent()),
		FHTTPTarget(r.RequestURI),
	}

	if r.Host != "" {
		fields = append(fields, FHTTPHost(r.Host))
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		fields = append(fields, FHTTPRequestID(id))
	}
	if id := r.Header.Get("X-Session-ID"); id != "" {
		fields = append(fields, FHTTPSessionID(id))
	}
	if r.TLS != nil {
		fields = append(fields, FHTTPScheme("https"))
	} else {
		fields = append(fields, FHTTPScheme("http"))
	}
	if r.ProtoMajor == 1 {
		fields = append(fields, FHTTPFlavor(fmt.Sprintf("1.%d", r.ProtoMinor)))
	} else if r.ProtoMajor == 2 {
		fields = append(fields, FHTTPFlavor("2"))
	}

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
		fields = append(fields, FHTTPClientIP(clientIP))
	}

	if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
		fields = append(fields, FTraceID(spanCtx.TraceID().String()))
	}

	return With(l, fields...)
}

func WithHTTPRequestOut(l *zap.Logger, r *http.Request) *zap.Logger {
	fields := []zap.Field{
		FHTTPWroteBytes(r.ContentLength),
		FHTTPMethod(r.Method),
		FHTTPTarget(r.URL.Path),
	}

	if r.URL.Host != "" {
		fields = append(fields, FHTTPHost(r.URL.Host))
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		fields = append(fields, FHTTPRequestID(id))
	}
	if id := r.Header.Get("X-Session-ID"); id != "" {
		fields = append(fields, FHTTPSessionID(id))
	}
	if r.TLS != nil {
		fields = append(fields, FHTTPScheme("https"))
	} else {
		fields = append(fields, FHTTPScheme("http"))
	}
	if r.ProtoMajor == 1 {
		fields = append(fields, FHTTPFlavor(fmt.Sprintf("1.%d", r.ProtoMinor)))
	} else if r.ProtoMajor == 2 {
		fields = append(fields, FHTTPFlavor("2"))
	}

	if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
		fields = append(fields, FTraceID(spanCtx.TraceID().String()))
	}

	return With(l, fields...)
}
