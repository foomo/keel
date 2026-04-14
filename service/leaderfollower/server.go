package leaderfollower

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"go.uber.org/zap"
)

// Server handles the /coord HTTP endpoint for peer-to-peer coordination.
// It extracts the action from the JSON envelope and delegates to [Protocol.Handle].
type Server struct {
	l        *zap.Logger
	protocol Protocol
	addr     string
	path     string
	srv      *http.Server
}

// ServerOption configures a [Server].
type ServerOption func(*Server)

// WithServerCoordPath sets the path for the coordination endpoint.
// Default: "/coord".
func WithServerCoordPath(path string) ServerOption {
	return func(s *Server) { s.path = path }
}

// NewServer creates a Server that listens on addr.
func NewServer(l *zap.Logger, protocol Protocol, addr string, opts ...ServerOption) *Server {
	s := &Server{
		l:        l,
		protocol: protocol,
		addr:     addr,
		path:     defaultCoordPath,
	}
	for _, o := range opts {
		o(s)
	}

	return s
}

// The handler responds to POST requests at any path (routing is the caller's job).
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.handle)
}

// Name implements the keel Service interface.
func (s *Server) Name() string { return "leaderfollower-server" }

// Start registers the /coord route and starts listening.
// Blocks until the server shuts down; returns nil on clean shutdown.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("POST "+s.path, s.Handler())

	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	s.l.Info("starting coord server", zap.String("addr", s.addr), zap.String("path", s.path))

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Close gracefully shuts down the HTTP server.
func (s *Server) Close(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}

	return nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var env FanOutRequest
	if err := json.Unmarshal(body, &env); err != nil {
		http.Error(w, "invalid envelope: "+err.Error(), http.StatusBadRequest)
		return
	}

	if env.Action == "" {
		http.Error(w, "action is required", http.StatusBadRequest)
		return
	}

	tw := &trackedWriter{ResponseWriter: w}
	if err := s.protocol.Handle(r.Context(), tw, env.Action, env.Payload); err != nil {
		s.l.Error("protocol handle error", zap.String("action", env.Action), zap.Error(err))

		if !tw.written {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// trackedWriter wraps http.ResponseWriter and records whether a response has
// been started, so the caller can avoid writing headers a second time.
type trackedWriter struct {
	http.ResponseWriter
	written bool
}

func (w *trackedWriter) WriteHeader(code int) {
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *trackedWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.ResponseWriter.Write(b)
}
