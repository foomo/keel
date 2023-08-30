package log

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

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

func WithError(l *zap.Logger, err error) *zap.Logger {
	return With(l, FErrorType(err), FError(err))
}

func WithServiceName(l *zap.Logger, name string) *zap.Logger {
	return With(l, FServiceName(name))
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
	if r.ProtoMajor == 1 {
		return With(l, FHTTPFlavor(fmt.Sprintf("1.%d", r.ProtoMinor)))
	} else if r.ProtoMajor == 2 {
		return With(l, FHTTPFlavor("2"))
	} else {
		return l
	}
}

func WithHTTPScheme(l *zap.Logger, r *http.Request) *zap.Logger {
	if r.TLS != nil {
		return With(l, FHTTPScheme("https"))
	} else {
		return With(l, FHTTPScheme("http"))
	}
}

func WithHTTPSessionID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Session-ID"); id != "" {
		return With(l, FHTTPSessionID(id))
	} else if id, ok := keelhttpcontext.GetSessionID(r.Context()); ok && id != "" {
		return With(l, FHTTPSessionID(id))
	} else {
		return l
	}
}

func WithHTTPRequestID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return With(l, FHTTPRequestID(id))
	} else if id, ok := keelhttpcontext.GetRequestID(r.Context()); ok && id != "" {
		return With(l, FHTTPRequestID(id))
	} else {
		return l
	}
}

func WithHTTPReferer(l *zap.Logger, r *http.Request) *zap.Logger {
	if value := r.Header.Get("X-Referer"); value != "" {
		return With(l, FHTTPReferer(value))
	} else if value := r.Referer(); value != "" {
		return With(l, FHTTPReferer(value))
	} else {
		return l
	}
}

func WithHTTPHost(l *zap.Logger, r *http.Request) *zap.Logger {
	if value := r.Header.Get("X-Forwarded-Host"); value != "" {
		return With(l, FHTTPHost(value))
	} else if !r.URL.IsAbs() {
		return With(l, FHTTPHost(r.Host))
	} else {
		return With(l, FHTTPHost(r.URL.Host))
	}
}

func WithHTTPTrackingID(l *zap.Logger, r *http.Request) *zap.Logger {
	if id := r.Header.Get("X-Tracking-ID"); id != "" {
		return With(l, FHTTPTrackingID(id))
	} else if id, ok := keelhttpcontext.GetTrackingID(r.Context()); ok && id != "" {
		return With(l, FHTTPTrackingID(id))
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
		return With(l, FHTTPClientIP(clientIP))
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
	return With(l,
		FHTTPMethod(r.Method),
		FHTTPTarget(r.RequestURI),
		FHTTPUserAgent(r.UserAgent()),
		FHTTPRequestContentLength(r.ContentLength),
	)
}

func WithHTTPRequestOut(l *zap.Logger, r *http.Request) *zap.Logger {
	l = WithHTTPHost(l, r)
	l = WithHTTPRequestID(l, r)
	l = WithHTTPSessionID(l, r)
	l = WithHTTPTrackingID(l, r)
	l = WithHTTPScheme(l, r)
	l = WithHTTPFlavor(l, r)
	l = WithTraceID(l, r.Context())
	return With(l,
		FHTTPMethod(r.Method),
		FHTTPTarget(r.URL.Path),
		FHTTPWroteBytes(r.ContentLength),
	)
}
