package log

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"
)

const (
	// MessagingSystemKey is the log field key "messaging_system".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingSystemKey = "messaging_system"
	// MessagingDestinationKey is the log field key "messaging_destination".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingDestinationKey = "messaging_destination"
	// MessagingDestinationKindKey is the log field key "messaging_destination_kind".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingDestinationKindKey = "messaging_destination_kind"
	// MessagingProtocolKey is the log field key "messaging_protocol".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingProtocolKey = "messaging_protocol"
	// MessagingProtocolVersionKey is the log field key "messaging_protocol_version".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingProtocolVersionKey = "messaging_protocol_version"
	// MessagingURLKey is the log field key "messaging_url".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingURLKey = "messaging_url"
	// MessagingMessageIDKey is the log field key "messaging_message_id".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingMessageIDKey = "messaging_message_id"
	// MessagingConversationIDKey is the log field key "messaging_conversation_id".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingConversationIDKey = "messaging_conversation_id"
	// MessagingMessagePayloadSizeBytesKey is the log field key "messaging_message_payload_size_bytes".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingMessagePayloadSizeBytesKey = "messaging_message_payload_size_bytes"
	// MessagingMessagePayloadCompressedSizeBytesKey is the log field key "messaging_message_payload_compressed_size_bytes".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingMessagePayloadCompressedSizeBytesKey = "messaging_message_payload_compressed_size_bytes"
)

// MessagingDestinationKind is the kind of a messaging destination.
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
type MessagingDestinationKind string

const (
	// MessagingDestinationKindQueue is the queue destination kind.
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingDestinationKindQueue MessagingDestinationKind = "queue"
	// MessagingDestinationKindTopic is the topic destination kind.
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	MessagingDestinationKindTopic MessagingDestinationKind = "topic"
)

// FMessagingSystem returns a field with the given value under [MessagingSystemKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingSystem(value string) zap.Field {
	return zap.String(MessagingSystemKey, value)
}

// FMessagingDestination returns the semconv.MessagingDestinationName attribute as a field.
//
// Deprecated: Use semconv.MessagingDestinationName with [Attribute] instead.
func FMessagingDestination(value string) zap.Field {
	return Attribute(semconv.MessagingDestinationName(value))
}

// FMessagingDestinationKind returns a field with the given value under [MessagingDestinationKindKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingDestinationKind(value MessagingDestinationKind) zap.Field {
	return zap.String(MessagingDestinationKindKey, string(value))
}

// FMessagingProtocol returns a field with the given value under [MessagingProtocolKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingProtocol(value string) zap.Field {
	return zap.String(MessagingProtocolKey, value)
}

// FMessagingProtocolVersion returns a field with the given value under [MessagingProtocolVersionKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingProtocolVersion(value string) zap.Field {
	return zap.String(MessagingProtocolVersionKey, value)
}

// FMessagingURL returns a field with the given value under [MessagingURLKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingURL(value string) zap.Field {
	return zap.String(MessagingURLKey, value)
}

// FMessagingMessageID returns the semconv.MessagingMessageID attribute as a field.
//
// Deprecated: Use semconv.MessagingMessageID with [Attribute] instead.
func FMessagingMessageID(value string) zap.Field {
	return Attribute(semconv.MessagingMessageID(value))
}

// FMessagingConversationID returns the semconv.MessagingMessageConversationID attribute as a field.
//
// Deprecated: Use semconv.MessagingMessageConversationID with [Attribute] instead.
func FMessagingConversationID(value string) zap.Field {
	return Attribute(semconv.MessagingMessageConversationID(value))
}

// FMessagingMessagePayloadSizeBytes returns a field with the given value under [MessagingMessagePayloadSizeBytesKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingMessagePayloadSizeBytes(value string) zap.Field {
	return zap.String(MessagingMessagePayloadSizeBytesKey, value)
}

// FMessagingMessagePayloadCompressedSizeBytes returns a field with the given value under [MessagingMessagePayloadCompressedSizeBytesKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FMessagingMessagePayloadCompressedSizeBytes(value string) zap.Field {
	return zap.String(MessagingMessagePayloadCompressedSizeBytesKey, value)
}
