package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/cookie"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	// TrackingIDOptions configures the [TrackingID] middleware.
	TrackingIDOptions struct {
		// Header is the request header the tracking ID is read from and, with
		// SetHeader, written to.
		Header string
		// Cookie is the cookie the tracking ID is read from and, with SetCookie,
		// written to.
		Cookie cookie.Cookie
		// Generator creates a tracking ID when SetCookie is enabled and the
		// request carries none.
		Generator TrackingIDGenerator
		// SetCookie reports whether a tracking ID is generated and set as cookie
		// when the request carries neither header nor cookie.
		SetCookie bool
		// SetHeader reports whether the tracking ID is set as request header.
		SetHeader bool
		// SetContext reports whether the tracking ID is stored in the request
		// context.
		SetContext bool
	}
	// TrackingIDOption configures [TrackingIDOptions].
	TrackingIDOption func(*TrackingIDOptions)
	// TrackingIDGenerator returns a new tracking ID.
	TrackingIDGenerator func() string
)

const (
	// DefaultTrackingIDCookieName is the default name of the tracking ID cookie.
	DefaultTrackingIDCookieName = "keel-tracking"
)

// DefaultTrackingIDGenerator returns a random UUID string.
func DefaultTrackingIDGenerator() string {
	return uuid.New().String()
}

// GetDefaultTrackingIDOptions returns the default options: the X-Tracking-ID header, a
// cookie named "keel-tracking", UUID generation, SetHeader and SetContext enabled
// and SetCookie disabled.
func GetDefaultTrackingIDOptions() TrackingIDOptions {
	return TrackingIDOptions{
		Header:     keelhttp.HeaderXTrackingID,
		Cookie:     cookie.New(DefaultTrackingIDCookieName),
		Generator:  DefaultTrackingIDGenerator,
		SetCookie:  false,
		SetHeader:  true,
		SetContext: true,
	}
}

// TrackingIDWithHeader sets [TrackingIDOptions.Header]. Defaults to X-Tracking-ID.
func TrackingIDWithHeader(v string) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Header = v
	}
}

// TrackingIDWithSetCookie sets [TrackingIDOptions.SetCookie]. Defaults to false.
func TrackingIDWithSetCookie(v bool) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.SetCookie = v
	}
}

// TrackingIDWithSetHeader sets [TrackingIDOptions.SetHeader]. Defaults to true.
func TrackingIDWithSetHeader(v bool) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.SetHeader = v
	}
}

// TrackingIDWithSetContext sets [TrackingIDOptions.SetContext]. Defaults to true.
func TrackingIDWithSetContext(v bool) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.SetContext = v
	}
}

// TrackingIDWithCookie sets [TrackingIDOptions.Cookie].
func TrackingIDWithCookie(v cookie.Cookie) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Cookie = v
	}
}

// TrackingIDWithGenerator sets [TrackingIDOptions.Generator]. Defaults to
// [DefaultTrackingIDGenerator].
func TrackingIDWithGenerator(v TrackingIDGenerator) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Generator = v
	}
}

// TrackingID returns a middleware that resolves the tracking ID from the request
// header or, failing that, the cookie. With SetCookie, a missing cookie is
// generated and set on both response and request. Cookie errors are answered
// with 500 Internal Server Error.
func TrackingID(opts ...TrackingIDOption) keelhttp.Middleware {
	options := GetDefaultTrackingIDOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return TrackingIDWithOptions(options)
}

// TrackingIDWithOptions is like [TrackingID] but takes fully populated options.
func TrackingIDWithOptions(opts TrackingIDOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("TrackingID")
			}

			var tackingID string
			if value := r.Header.Get(opts.Header); value != "" {
				tackingID = value
			} else if c, err := opts.Cookie.Get(r); errors.Is(err, http.ErrNoCookie) && !opts.SetCookie {
				// do nothing
			} else if errors.Is(err, http.ErrNoCookie) && opts.SetCookie {
				tackingID = opts.Generator()
				if c, err := opts.Cookie.Set(w, r, tackingID); err != nil {
					httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to set tracking id cookie"))
					return
				} else {
					r.AddCookie(c)
				}
			} else if err != nil {
				httputils.InternalServerError(l, w, r, errors.Wrap(err, "failed to read tracking id cookie"))
				return
			} else {
				tackingID = c.Value
			}

			if span.IsRecording() && tackingID != "" {
				span.SetAttributes(semconv.HTTPRequestHeader(strings.ToLower(opts.Header), tackingID))
			}

			if tackingID != "" && opts.SetHeader {
				r.Header.Set(opts.Header, tackingID)
			}

			if tackingID != "" && opts.SetContext {
				r = r.WithContext(keelhttpcontext.SetTrackingID(r.Context(), tackingID))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TrackingIDFromContext returns the tracking ID stored in ctx, or "" if there is none.
func TrackingIDFromContext(ctx context.Context) string {
	if value, ok := keelhttpcontext.GetTrackingID(ctx); ok {
		return value
	}

	return ""
}
