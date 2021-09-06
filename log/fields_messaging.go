package log

import (
	"go.uber.org/zap"
)

const (
	MessagingSystemKey                            = "messaging_system"
	MessagingDestinationKey                       = "messaging_destination"
	MessagingDestinationKindKey                   = "messaging_destination_kind"
	MessagingProtocolKey                          = "messaging_protocol"
	MessagingProtocolVersionKey                   = "messaging_protocol_version"
	MessagingURLKey                               = "messaging_url"
	MessagingMessageIDKey                         = "messaging_message_id"
	MessagingConversationIDKey                    = "messaging_conversation_id"
	MessagingMessagePayloadSizeBytesKey           = "messaging_message_payload_size_bytes"
	MessagingMessagePayloadCompressedSizeBytesKey = "messaging_message_payload_compressed_size_bytes"
)

type MessagingDestinationKind string

const (
	MessagingDestinationKindQueue MessagingDestinationKind = "queue"
	MessagingDestinationKindTopic MessagingDestinationKind = "topic"
)

func FMessagingSystem(value string) zap.Field {
	return zap.String(MessagingSystemKey, value)
}

func FMessagingDestination(value string) zap.Field {
	return zap.String(MessagingDestinationKey, value)
}

func FMessagingDestinationKind(value MessagingDestinationKind) zap.Field {
	return zap.String(MessagingDestinationKindKey, string(value))
}

func FMessagingProtocol(value string) zap.Field {
	return zap.String(MessagingProtocolKey, value)
}

func FMessagingProtocolVersion(value string) zap.Field {
	return zap.String(MessagingProtocolVersionKey, value)
}

func FMessagingURL(value string) zap.Field {
	return zap.String(MessagingURLKey, value)
}

func FMessagingMessageID(value string) zap.Field {
	return zap.String(MessagingMessageIDKey, value)
}

func FMessagingConversationID(value string) zap.Field {
	return zap.String(MessagingConversationIDKey, value)
}

func FMessagingMessagePayloadSizeBytes(value string) zap.Field {
	return zap.String(MessagingMessagePayloadSizeBytesKey, value)
}

func FMessagingMessagePayloadCompressedSizeBytes(value string) zap.Field {
	return zap.String(MessagingMessagePayloadCompressedSizeBytesKey, value)
}
