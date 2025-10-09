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

func GoTSRPCFunc(v string) attribute.KeyValue {
	return GoTSRPCFuncKey.String(v)
}

func GoTSRPCService(v string) attribute.KeyValue {
	return GoTSRPCServiceKey.String(v)
}

func GoTSRPCPackage(v string) attribute.KeyValue {
	return GoTSRPCPackageKey.String(v)
}

func GoTSRPCMarshalling(v int64) attribute.KeyValue {
	return GoTSRPCMarshallingKey.Int64(v)
}

func GoTSRPCUnmarshalling(v int64) attribute.KeyValue {
	return GoTSRPCUnmarshallingKey.Int64(v)
}

func GoTSRPCPayload(v string) attribute.KeyValue {
	return GoTSRPCPayloadKey.String(v)
}

func GoTSRPCErrorCode(v int) attribute.KeyValue {
	return GoTSRPCErrorCodeKey.Int(v)
}

func GoTSRPCErrorMessage(v string) attribute.KeyValue {
	return GoTSRPCErrorMessageKey.String(v)
}

func GoTSRPCErrorType(v string) attribute.KeyValue {
	return GoTSRPCErrorTypeKey.String(v)
}
