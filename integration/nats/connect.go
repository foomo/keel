package nats

import (
	"github.com/foomo/keel"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Connect establishes a connection to a NATS server with optional OTel instrumentation.
func Connect(s *keel.Server, url string, opts ...nats.Option) (*nats.Conn, error) {
	l := s.Logger()

	opts = append(opts,
		nats.ConnectHandler(func(conn *nats.Conn) {
			l.Debug("connected",
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			l.Warn("disconnected", zap.Error(err),
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),
		nats.ReconnectErrHandler(func(conn *nats.Conn, err error) {
			l.Warn("reconnect", zap.Error(err),
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			l.Info("closed",
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),
		nats.NoCallbacksAfterClientClose(),
	)

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	s.AddCloser(conn)

	return conn, nil
}
