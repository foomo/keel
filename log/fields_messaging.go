package log

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
)

const (
	// Deprecated: use semconv messaging attributes instead.
	MessagingSystemKey = "messaging_system"
	// Deprecated: use semconv messaging attributes instead.
	MessagingDestinationKey = "messaging_destination"
	// Deprecated: use semconv messaging attributes instead.
	MessagingDestinationKindKey = "messaging_destination_kind"
	// Deprecated: use semconv messaging attributes instead.
	MessagingProtocolKey = "messaging_protocol"
	// Deprecated: use semconv messaging attributes instead.
	MessagingProtocolVersionKey = "messaging_protocol_version"
	// Deprecated: use semconv messaging attributes instead.
	MessagingURLKey = "messaging_url"
	// Deprecated: use semconv messaging attributes instead.
	MessagingMessageIDKey = "messaging_message_id"
	// Deprecated: use semconv messaging attributes instead.
	MessagingConversationIDKey = "messaging_conversation_id"
	// Deprecated: use semconv messaging attributes instead.
	MessagingMessagePayloadSizeBytesKey = "messaging_message_payload_size_bytes"
	// Deprecated: use semconv messaging attributes instead.
	MessagingMessagePayloadCompressedSizeBytesKey = "messaging_message_payload_compressed_size_bytes"
)

// Deprecated: use semconv messaging attributes instead.
type MessagingDestinationKind string

const (
	// Deprecated: use semconv messaging attributes instead.
	MessagingDestinationKindQueue MessagingDestinationKind = "queue"
	// Deprecated: use semconv messaging attributes instead.
	MessagingDestinationKindTopic MessagingDestinationKind = "topic"
)

// Deprecated: use semconv messaging attributes instead.
func FMessagingSystem(value string) zap.Field {
	return zap.String(MessagingSystemKey, value)
}

// Deprecated: use semconv.MessagingDestinationName instead.
func FMessagingDestination(value string) zap.Field {
	return Attribute(semconv.MessagingDestinationName(value))
}

// Deprecated: use semconv messaging attributes instead.
func FMessagingDestinationKind(value MessagingDestinationKind) zap.Field {
	return zap.String(MessagingDestinationKindKey, string(value))
}

// Deprecated: use semconv messaging attributes instead.
func FMessagingProtocol(value string) zap.Field {
	return zap.String(MessagingProtocolKey, value)
}

// Deprecated: use semconv messaging attributes instead.
func FMessagingProtocolVersion(value string) zap.Field {
	return zap.String(MessagingProtocolVersionKey, value)
}

// Deprecated: use semconv messaging attributes instead.
func FMessagingURL(value string) zap.Field {
	return zap.String(MessagingURLKey, value)
}

// Deprecated: use semconv.MessagingMessageID instead.
func FMessagingMessageID(value string) zap.Field {
	return Attribute(semconv.MessagingMessageID(value))
}

// Deprecated: use semconv.MessagingMessageConversationID instead.
func FMessagingConversationID(value string) zap.Field {
	return Attribute(semconv.MessagingMessageConversationID(value))
}

// Deprecated: use semconv messaging attributes instead.
func FMessagingMessagePayloadSizeBytes(value string) zap.Field {
	return zap.String(MessagingMessagePayloadSizeBytesKey, value)
}

// Deprecated: use semconv messaging attributes instead.
func FMessagingMessagePayloadCompressedSizeBytes(value string) zap.Field {
	return zap.String(MessagingMessagePayloadCompressedSizeBytesKey, value)
}
