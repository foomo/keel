package keel

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

const (
	DefaultServiceHTTPZapName = "zap"
	DefaultServiceHTTPZapAddr = "localhost:9100"
	DefaultServiceHTTPZapPath = "/log"
)

func NewServiceHTTPZap(l *zap.Logger, addr, path string) *ServiceHTTP {
	handler := http.NewServeMux()
	handler.Handle(path, zap.NewAtomicLevel())
	return NewServiceHTTP(l, addr, handler)
}

func NewDefaultServiceHTTPZap() *ServiceHTTP {
	return NewServiceHTTPZap(
		log.Logger().With(log.FServiceName(DefaultServiceHTTPZapName)),
		DefaultServiceHTTPZapAddr,
		DefaultServiceHTTPZapPath,
	)
}
