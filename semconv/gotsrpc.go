package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	GoTSRPCFuncKey          = attribute.Key("gotsrpc.func")
	GoTSRPCServiceKey       = attribute.Key("gotsrpc.service")
	GoTSRPCPackageKey       = attribute.Key("gotsrpc.package")
	GoTSRPCMarshallingKey   = attribute.Key("gotsrpc.marshalling")
	GoTSRPCUnmarshallingKey = attribute.Key("gotsrpc.unmarshalling")
	GoTSRPCPayloadKey       = attribute.Key("gotsrpc.payload")
	GoTSRPCErrorCodeKey     = attribute.Key("gotsrpc.error.code")
	GoTSRPCErrorMessageKey  = attribute.Key("gotsrpc.error.message")
	GoTSRPCErrorTypeKey     = attribute.Key("gotsrpc.error.type")
)

// GoTSRPCFunc returns a new attribute.KeyValue for gotsrpc.func.
func GoTSRPCFunc(v string) attribute.KeyValue {
	return GoTSRPCFuncKey.String(v)
}

// GoTSRPCService returns a new attribute.KeyValue for gotsrpc.service.
func GoTSRPCService(v string) attribute.KeyValue {
	return GoTSRPCServiceKey.String(v)
}

// GoTSRPCPackage returns a new attribute.KeyValue for gotsrpc.package.
func GoTSRPCPackage(v string) attribute.KeyValue {
	return GoTSRPCPackageKey.String(v)
}

// GoTSRPCMarshalling returns a new attribute.KeyValue for gotsrpc.marshalling.
func GoTSRPCMarshalling(v int64) attribute.KeyValue {
	return GoTSRPCMarshallingKey.Int64(v)
}

// GoTSRPCUnmarshalling returns a new attribute.KeyValue for gotsrpc.unmarshalling.
func GoTSRPCUnmarshalling(v int64) attribute.KeyValue {
	return GoTSRPCUnmarshallingKey.Int64(v)
}

// GoTSRPCPayload returns a new attribute.KeyValue for gotsrpc.payload.
func GoTSRPCPayload(v string) attribute.KeyValue {
	return GoTSRPCPayloadKey.String(v)
}

// GoTSRPCErrorCode returns a new attribute.KeyValue for gotsrpc.error.code.
func GoTSRPCErrorCode(v int) attribute.KeyValue {
	return GoTSRPCErrorCodeKey.Int(v)
}

// GoTSRPCErrorMessage returns a new attribute.KeyValue for gotsrpc.error.message.
func GoTSRPCErrorMessage(v string) attribute.KeyValue {
	return GoTSRPCErrorMessageKey.String(v)
}

// GoTSRPCErrorType returns a new attribute.KeyValue for gotsrpc.error.type.
func GoTSRPCErrorType(v string) attribute.KeyValue {
	return GoTSRPCErrorTypeKey.String(v)
}
