package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/cookie"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	TrackingIDOptions struct {
		// Header to look up the tracking id
		Header string
		// Cookie how to set the cookie
		Cookie cookie.Cookie
		// Generator for the tracking ids
		Generator TrackingIDGenerator
		// SetCookie if true it will create a cookie if not exists
		SetCookie bool
		// SetHeader if true it will set add a request header
		SetHeader bool
		// SetContext if true it will set the context key
		SetContext bool
	}
	TrackingIDOption    func(*TrackingIDOptions)
	TrackingIDGenerator func() string
)

const (
	DefaultTrackingIDCookieName = "keel-tracking"
)

// DefaultTrackingIDGenerator function
func DefaultTrackingIDGenerator() string {
	return uuid.New().String()
}

// GetDefaultTrackingIDOptions returns the default options
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

func TrackingIDWithHeader(v string) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Header = v
	}
}

// TrackingIDWithSetCookie middleware option
func TrackingIDWithSetCookie(v bool) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.SetCookie = v
	}
}

// TrackingIDWithSetHeader middleware option
func TrackingIDWithSetHeader(v bool) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.SetHeader = v
	}
}

// TrackingIDWithSetContext middleware option
func TrackingIDWithSetContext(v bool) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.SetContext = v
	}
}

// TrackingIDWithCookie middleware option
func TrackingIDWithCookie(v cookie.Cookie) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Cookie = v
	}
}

// TrackingIDWithGenerator middleware option
func TrackingIDWithGenerator(v TrackingIDGenerator) TrackingIDOption {
	return func(o *TrackingIDOptions) {
		o.Generator = v
	}
}

// TrackingID middleware
func TrackingID(opts ...TrackingIDOption) Middleware {
	options := GetDefaultTrackingIDOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return TrackingIDWithOptions(options)
}

// TrackingIDWithOptions middleware
func TrackingIDWithOptions(opts TrackingIDOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// TrackingIDFromContext helper
func TrackingIDFromContext(ctx context.Context) string {
	if value, ok := keelhttpcontext.GetTrackingID(ctx); ok {
		return value
	}
	return ""
}
