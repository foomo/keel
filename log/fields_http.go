package log

import (
	"go.uber.org/zap"
)

const (
	// HTTPMethodKey represents the HTTP request method.
	HTTPMethodKey = "http_method"

	// HTTPTargetKey represents the full request target as passed in a HTTP
	// request line or equivalent, e.g. "/path/12314/?q=ddds#123".
	HTTPTargetKey = "http_target"

	// HTTPHostKey represents the value of the HTTP host header.
	HTTPHostKey = "http_host"

	// HTTPStatusCodeKey represents the HTTP response status code.
	HTTPStatusCodeKey = "http_status_code"

	// HTTPUserAgentKey represents the Value of the HTTP User-Agent header sent by the client.
	HTTPUserAgentKey = "http_user_agent"

	// HTTPClientIPKey represents the IP address of the original client behind all proxies,
	// if known (e.g. from X-Forwarded-For).
	HTTPClientIPKey = "http_client_ip"

	// HTTPRequestContentLengthKey represents the size of the request payload body in bytes.
	HTTPRequestContentLengthKey = "http_read_bytes"

	// HTTPWroteBytesKey represents the size of the response payload body in bytes.
	HTTPWroteBytesKey = "http_wrote_bytes"

	// HTTPSchemeKey represents the URI scheme identifying the used protocol.
	HTTPSchemeKey = "http_scheme"

	// HTTPFlavorKey represents the Kind of HTTP protocol used.
	HTTPFlavorKey = "http_flavor"

	// HTTPRequestIDKey represents the HTTP request id if known (e.g from X-Request-ID).
	HTTPRequestIDKey = "http_request_id"

	// HTTPSessionIDKey represents the HTTP session id if known (e.g from X-Session-ID).
	HTTPSessionIDKey = "http_session_id"
)

func FHTTPRequestID(id string) zap.Field {
	return zap.String(HTTPRequestIDKey, id)
}

func FHTTPSessionID(id string) zap.Field {
	return zap.String(HTTPSessionIDKey, id)
}

func FHTTPRequestContentLength(bytes int64) zap.Field {
	return zap.Int64(HTTPRequestContentLengthKey, bytes)
}

func FHTTPWroteBytes(bytes int64) zap.Field {
	return zap.Int64(HTTPWroteBytesKey, bytes)
}

func FHTTPStatusCode(status int) zap.Field {
	return zap.Int(HTTPStatusCodeKey, status)
}

func FHTTPTarget(target string) zap.Field {
	return zap.String(HTTPTargetKey, target)
}

func FHTTPClientIP(clientIP string) zap.Field {
	return zap.String(HTTPClientIPKey, clientIP)
}

func FHTTPFlavor(flavor string) zap.Field {
	return zap.String(HTTPFlavorKey, flavor)
}

func FHTTPScheme(scheme string) zap.Field {
	return zap.String(HTTPSchemeKey, scheme)
}

func FHTTPUserAgent(userAgent string) zap.Field {
	return zap.String(HTTPUserAgentKey, userAgent)
}

func FHTTPHost(host string) zap.Field {
	return zap.String(HTTPHostKey, host)
}

func FHTTPMethod(name string) zap.Field {
	return zap.String(HTTPMethodKey, name)
}
