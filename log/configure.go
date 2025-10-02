package log

import (
	"net/http"

	"go.uber.org/zap"
)

type Config struct {
	l *zap.Logger
}

func Configure(l *zap.Logger) *Config {
	if l == nil {
		l = Logger()
	}

	return &Config{l: l.With()}
}

func (c *Config) Logger() *zap.Logger {
	return c.l
}

func (c *Config) Error(err error) *Config {
	c.l = c.l.With(FError(err))
	return c
}

func (c *Config) With(fields ...zap.Field) *Config {
	c.l = c.l.With(fields...)
	return c
}

func (c *Config) HTTPRequest(r *http.Request) *Config {
	c.l = WithHTTPRequest(c.l, r)
	return c
}
