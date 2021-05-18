package log

import (
	"go.uber.org/zap"
)

const (
	// NetHostIPKey represents the local host IP. Useful in case of a multi-IP host.
	NetHostIPKey = "net_host_ip"

	// NetHostPortKey represents the local host port.
	NetHostPortKey = "net_host_port"
)

func FNetHostIP(ip string) zap.Field {
	return zap.String(NetHostIPKey, ip)
}

func FNetHostPort(port string) zap.Field {
	return zap.String(NetHostPortKey, port)
}
