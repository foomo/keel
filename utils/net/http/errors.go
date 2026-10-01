package httputils

import (
	"errors"
	"net/http"

	httplog "github.com/foomo/keel/net/http/log"
	"github.com/foomo/keel/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// InternalServerError responds with status [http.StatusInternalServerError] using [ServerError].
func InternalServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusInternalServerError, errs...)
}

// InternalServiceUnavailable responds with status [http.StatusServiceUnavailable] using [ServerError].
func InternalServiceUnavailable(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusServiceUnavailable, errs...)
}

// InternalServiceTooEarly responds with status [http.StatusTooEarly] using [ServerError].
func InternalServiceTooEarly(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusTooEarly, errs...)
}

// UnauthorizedServerError responds with status [http.StatusUnauthorized] using [ServerError].
func UnauthorizedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusUnauthorized, errs...)
}

// BadRequestServerError responds with status [http.StatusBadRequest] using [ServerError].
func BadRequestServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusBadRequest, errs...)
}

// MethodNotAllowedServerError responds with status [http.StatusMethodNotAllowed] using [ServerError].
func MethodNotAllowedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusMethodNotAllowed, errs...)
}

// RequestEntityTooLargeServerError responds with status [http.StatusRequestEntityTooLarge] using [ServerError].
func RequestEntityTooLargeServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusRequestEntityTooLarge, errs...)
}

// NotFoundServerError responds with status [http.StatusNotFound] using [ServerError].
func NotFoundServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusNotFound, errs...)
}

// PaymentRequiredServerError responds with status [http.StatusPaymentRequired] using [ServerError].
func PaymentRequiredServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusPaymentRequired, errs...)
}

// ForbiddenServerError responds with status [http.StatusForbidden] using [ServerError].
func ForbiddenServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusForbidden, errs...)
}

// NotAcceptableServerError responds with status [http.StatusNotAcceptable] using [ServerError].
func NotAcceptableServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusNotAcceptable, errs...)
}

// ProxyAuthRequiredServerError responds with status [http.StatusProxyAuthRequired] using [ServerError].
func ProxyAuthRequiredServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusProxyAuthRequired, errs...)
}

// RequestTimeoutServerError responds with status [http.StatusRequestTimeout] using [ServerError].
func RequestTimeoutServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusRequestTimeout, errs...)
}

// ConflictServerError responds with status [http.StatusConflict] using [ServerError].
func ConflictServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusConflict, errs...)
}

// GoneServerError responds with status [http.StatusGone] using [ServerError].
func GoneServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusGone, errs...)
}

// LengthRequiredServerError responds with status [http.StatusLengthRequired] using [ServerError].
func LengthRequiredServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusLengthRequired, errs...)
}

// PreconditionFailedServerError responds with status [http.StatusPreconditionFailed] using [ServerError].
func PreconditionFailedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusPreconditionFailed, errs...)
}

// RequestURITooLongServerError responds with status [http.StatusRequestURITooLong] using [ServerError].
func RequestURITooLongServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusRequestURITooLong, errs...)
}

// UnsupportedMediaTypeServerError responds with status [http.StatusUnsupportedMediaType] using [ServerError].
func UnsupportedMediaTypeServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusUnsupportedMediaType, errs...)
}

// RequestedRangeNotSatisfiableServerError responds with status [http.StatusRequestedRangeNotSatisfiable] using [ServerError].
func RequestedRangeNotSatisfiableServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusRequestedRangeNotSatisfiable, errs...)
}

// ExpectationFailedServerError responds with status [http.StatusExpectationFailed] using [ServerError].
func ExpectationFailedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusExpectationFailed, errs...)
}

// TeapotServerError responds with status [http.StatusTeapot] using [ServerError].
func TeapotServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusTeapot, errs...)
}

// MisdirectedRequestServerError responds with status [http.StatusMisdirectedRequest] using [ServerError].
func MisdirectedRequestServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusMisdirectedRequest, errs...)
}

// UnprocessableEntityServerError responds with status [http.StatusUnprocessableEntity] using [ServerError].
func UnprocessableEntityServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusUnprocessableEntity, errs...)
}

// LockedServerError responds with status [http.StatusLocked] using [ServerError].
func LockedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusLocked, errs...)
}

// FailedDependencyServerError responds with status [http.StatusFailedDependency] using [ServerError].
func FailedDependencyServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusFailedDependency, errs...)
}

// UpgradeRequiredServerError responds with status [http.StatusUpgradeRequired] using [ServerError].
func UpgradeRequiredServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusUpgradeRequired, errs...)
}

// PreconditionRequiredServerError responds with status [http.StatusPreconditionRequired] using [ServerError].
func PreconditionRequiredServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusPreconditionRequired, errs...)
}

// TooManyRequestsServerError responds with status [http.StatusTooManyRequests] using [ServerError].
func TooManyRequestsServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusTooManyRequests, errs...)
}

// RequestHeaderFieldsTooLargeServerError responds with status [http.StatusRequestHeaderFieldsTooLarge] using [ServerError].
func RequestHeaderFieldsTooLargeServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusRequestHeaderFieldsTooLarge, errs...)
}

// UnavailableForLegalReasonsServerError responds with status [http.StatusUnavailableForLegalReasons] using [ServerError].
func UnavailableForLegalReasonsServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusUnavailableForLegalReasons, errs...)
}

// NotImplementedServerError responds with status [http.StatusNotImplemented] using [ServerError].
func NotImplementedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusNotImplemented, errs...)
}

// BadGatewayServerError responds with status [http.StatusBadGateway] using [ServerError].
func BadGatewayServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusBadGateway, errs...)
}

// GatewayTimeoutServerError responds with status [http.StatusGatewayTimeout] using [ServerError].
func GatewayTimeoutServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusGatewayTimeout, errs...)
}

// HTTPVersionNotSupportedServerError responds with status [http.StatusHTTPVersionNotSupported] using [ServerError].
func HTTPVersionNotSupportedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusHTTPVersionNotSupported, errs...)
}

// VariantAlsoNegotiatesServerError responds with status [http.StatusVariantAlsoNegotiates] using [ServerError].
func VariantAlsoNegotiatesServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusVariantAlsoNegotiates, errs...)
}

// InsufficientStorageServerError responds with status [http.StatusInsufficientStorage] using [ServerError].
func InsufficientStorageServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusInsufficientStorage, errs...)
}

// LoopDetectedServerError responds with status [http.StatusLoopDetected] using [ServerError].
func LoopDetectedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusLoopDetected, errs...)
}

// NotExtendedServerError responds with status [http.StatusNotExtended] using [ServerError].
func NotExtendedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusNotExtended, errs...)
}

// NetworkAuthenticationRequiredServerError responds with status [http.StatusNetworkAuthenticationRequired] using [ServerError].
func NetworkAuthenticationRequiredServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, errs ...error) {
	ServerError(l, w, r, http.StatusNetworkAuthenticationRequired, errs...)
}

// ServerError writes a plain-text response with the given status code and its
// [http.StatusText] as body, if errs joined by [errors.Join] is non-nil.
//
// The error is recorded on the current span and its type added to the
// otelhttp labeler, if any. If r carries a request log labeler (see
// [httplog.InjectLabelerIntoRequest]), the error is added to it; otherwise it
// is logged at error level through l, unless l is nil.
//
// When errs is empty or contains only nil errors, nothing is written to w.
func ServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, code int, errs ...error) {
	if err := errors.Join(errs...); err != nil {
		errType := semconv.ErrorType(err)
		telemetry.Ctx(r.Context()).RecordError(err)

		if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
			labeler.Add(errType)
		}

		// add log entry
		if labeler, ok := httplog.LabelerFromRequest(r); ok {
			labeler.Add(log.Attribute(errType), log.FError(err))
		} else if l != nil {
			l = log.WithError(l, err)
			l = log.WithHTTPRequest(l, r)
			l.Error("http server error", log.Attribute(semconv.HTTPResponseStatusCode(code)))
		}

		http.Error(w, http.StatusText(code), code)
	}
}
