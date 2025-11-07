package log

import (
	"go.uber.org/zap"
)

const (
	// Deprecated: use semconv messaging attributes instead.
	HTTPServerNameKey = "http_server_name"
	// Deprecated: use semconv messaging attributes instead.
	HTTPMethodKey = "http_method"
	// Deprecated: use semconv messaging attributes instead.
	HTTPTargetKey = "http_target"
	// Deprecated: use semconv messaging attributes instead.
	HTTPHostKey = "http_host"
	// Deprecated: use semconv messaging attributes instead.
	HTTPStatusCodeKey = "http_status_code"
	// Deprecated: use semconv messaging attributes instead.
	HTTPUserAgentKey = "http_user_agent"
	// Deprecated: use semconv messaging attributes instead.
	HTTPClientIPKey = "http_client_ip"
	// Deprecated: use semconv messaging attributes instead.
	HTTPRequestContentLengthKey = "http_read_bytes"
	// Deprecated: use semconv messaging attributes instead.
	HTTPWroteBytesKey = "http_wrote_bytes" // #nosec
	// Deprecated: use semconv messaging attributes instead.
	HTTPSchemeKey = "http_scheme"
	// Deprecated: use semconv messaging attributes instead.
	HTTPFlavorKey = "http_flavor"
	// Deprecated: use semconv messaging attributes instead.
	HTTPRequestIDKey = "http_request_id"
	// Deprecated: use semconv messaging attributes instead.
	HTTPSessionIDKey = "http_session_id"
	// Deprecated: use semconv messaging attributes instead.
	HTTPTrackingIDKey = "http_tracking_id"
	// Deprecated: use semconv messaging attributes instead.
	HTTPRefererKey = "http_referer"
)

// Deprecated: use semconv messaging attributes instead.
func FHTTPServerName(id string) zap.Field {
	return zap.String(HTTPServerNameKey, id)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPRequestID(id string) zap.Field {
	return zap.String(HTTPRequestIDKey, id)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPSessionID(id string) zap.Field {
	return zap.String(HTTPSessionIDKey, id)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPTrackingID(id string) zap.Field {
	return zap.String(HTTPTrackingIDKey, id)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPRequestContentLength(bytes int64) zap.Field {
	return zap.Int64(HTTPRequestContentLengthKey, bytes)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPWroteBytes(bytes int64) zap.Field {
	return zap.Int64(HTTPWroteBytesKey, bytes)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPStatusCode(status int) zap.Field {
	return zap.Int(HTTPStatusCodeKey, status)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPTarget(target string) zap.Field {
	return zap.String(HTTPTargetKey, target)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPClientIP(clientIP string) zap.Field {
	return zap.String(HTTPClientIPKey, clientIP)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPFlavor(flavor string) zap.Field {
	return zap.String(HTTPFlavorKey, flavor)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPScheme(scheme string) zap.Field {
	return zap.String(HTTPSchemeKey, scheme)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPUserAgent(userAgent string) zap.Field {
	return zap.String(HTTPUserAgentKey, userAgent)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPReferer(host string) zap.Field {
	return zap.String(HTTPRefererKey, host)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPHost(host string) zap.Field {
	return zap.String(HTTPHostKey, host)
}

// Deprecated: use semconv messaging attributes instead.
func FHTTPMethod(name string) zap.Field {
	return zap.String(HTTPMethodKey, name)
}
