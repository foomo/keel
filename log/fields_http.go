package log

import (
	"go.uber.org/zap"
)

const (
	// HTTPServerNameKey is the log field key "http_server_name".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPServerNameKey = "http_server_name"
	// HTTPMethodKey is the log field key "http_method".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPMethodKey = "http_method"
	// HTTPTargetKey is the log field key "http_target".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPTargetKey = "http_target"
	// HTTPHostKey is the log field key "http_host".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPHostKey = "http_host"
	// HTTPStatusCodeKey is the log field key "http_status_code".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPStatusCodeKey = "http_status_code"
	// HTTPUserAgentKey is the log field key "http_user_agent".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPUserAgentKey = "http_user_agent"
	// HTTPClientIPKey is the log field key "http_client_ip".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPClientIPKey = "http_client_ip"
	// HTTPRequestContentLengthKey is the log field key "http_read_bytes".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPRequestContentLengthKey = "http_read_bytes"
	// HTTPWroteBytesKey is the log field key "http_wrote_bytes".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPWroteBytesKey = "http_wrote_bytes" // #nosec
	// HTTPSchemeKey is the log field key "http_scheme".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPSchemeKey = "http_scheme"
	// HTTPFlavorKey is the log field key "http_flavor".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPFlavorKey = "http_flavor"
	// HTTPRequestIDKey is the log field key "http_request_id".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPRequestIDKey = "http_request_id"
	// HTTPSessionIDKey is the log field key "http_session_id".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPSessionIDKey = "http_session_id"
	// HTTPTrackingIDKey is the log field key "http_tracking_id".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPTrackingIDKey = "http_tracking_id"
	// HTTPRefererKey is the log field key "http_referer".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	HTTPRefererKey = "http_referer"
)

// FHTTPServerName returns a field with the given value under [HTTPServerNameKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPServerName(id string) zap.Field {
	return zap.String(HTTPServerNameKey, id)
}

// FHTTPRequestID returns a field with the given value under [HTTPRequestIDKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPRequestID(id string) zap.Field {
	return zap.String(HTTPRequestIDKey, id)
}

// FHTTPSessionID returns a field with the given value under [HTTPSessionIDKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPSessionID(id string) zap.Field {
	return zap.String(HTTPSessionIDKey, id)
}

// FHTTPTrackingID returns a field with the given value under [HTTPTrackingIDKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPTrackingID(id string) zap.Field {
	return zap.String(HTTPTrackingIDKey, id)
}

// FHTTPRequestContentLength returns a field with the given value under [HTTPRequestContentLengthKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPRequestContentLength(bytes int64) zap.Field {
	return zap.Int64(HTTPRequestContentLengthKey, bytes)
}

// FHTTPWroteBytes returns a field with the given value under [HTTPWroteBytesKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPWroteBytes(bytes int64) zap.Field {
	return zap.Int64(HTTPWroteBytesKey, bytes)
}

// FHTTPStatusCode returns a field with the given value under [HTTPStatusCodeKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPStatusCode(status int) zap.Field {
	return zap.Int(HTTPStatusCodeKey, status)
}

// FHTTPTarget returns a field with the given value under [HTTPTargetKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPTarget(target string) zap.Field {
	return zap.String(HTTPTargetKey, target)
}

// FHTTPClientIP returns a field with the given value under [HTTPClientIPKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPClientIP(clientIP string) zap.Field {
	return zap.String(HTTPClientIPKey, clientIP)
}

// FHTTPFlavor returns a field with the given value under [HTTPFlavorKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPFlavor(flavor string) zap.Field {
	return zap.String(HTTPFlavorKey, flavor)
}

// FHTTPScheme returns a field with the given value under [HTTPSchemeKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPScheme(scheme string) zap.Field {
	return zap.String(HTTPSchemeKey, scheme)
}

// FHTTPUserAgent returns a field with the given value under [HTTPUserAgentKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPUserAgent(userAgent string) zap.Field {
	return zap.String(HTTPUserAgentKey, userAgent)
}

// FHTTPReferer returns a field with the given value under [HTTPRefererKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPReferer(host string) zap.Field {
	return zap.String(HTTPRefererKey, host)
}

// FHTTPHost returns a field with the given value under [HTTPHostKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPHost(host string) zap.Field {
	return zap.String(HTTPHostKey, host)
}

// FHTTPMethod returns a field with the given value under [HTTPMethodKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FHTTPMethod(name string) zap.Field {
	return zap.String(HTTPMethodKey, name)
}
