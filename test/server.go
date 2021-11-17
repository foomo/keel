package keeltest

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
)

type Server struct {
	services   []Service
	serviceMap map[string]Service
	ctx        context.Context
	l          *zap.Logger
	c          *viper.Viper
}

func NewServer(opts ...Option) *Server {
	inst := &Server{
		ctx: context.Background(),
		c:   config.Config(),
		l:   log.Logger(),
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

// Logger returns server logger
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Config returns server config
func (s *Server) Config() *viper.Viper {
	return s.c
}

// Context returns server context
func (s *Server) Context() context.Context {
	return s.ctx
}

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	for _, value := range s.services {
		if value == service {
			return
		}
	}
	s.services = append(s.services, service)
}

func (s *Server) GetService(name string) Service {
	if v, ok := s.serviceMap[name]; ok {
		return v
	}
	return nil
}

// Start starts all registered services
func (s *Server) Start() {
	s.serviceMap = make(map[string]Service, len(s.services))
	for _, service := range s.services {
		s.serviceMap[service.Name()] = service
		service.Start(s.Context())
	}
}
