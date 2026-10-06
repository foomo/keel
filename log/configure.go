package log

import (
	"net/http"

	"go.uber.org/zap"
)

// Config is a builder that accumulates fields on a [zap.Logger]. Its methods
// modify the receiver and return it for chaining. Use [Configure] to create
// one.
type Config struct {
	l *zap.Logger
}

// Configure returns a [Config] starting from l, or from [Logger] if l is nil.
func Configure(l *zap.Logger) *Config {
	if l == nil {
		l = Logger()
	}

	return &Config{l: l.With()}
}

// Logger returns the configured logger.
func (c *Config) Logger() *zap.Logger {
	return c.l
}

// Error adds err using [WithError].
func (c *Config) Error(err error) *Config {
	c.l = WithError(c.l, err)
	return c
}

// With adds fields.
func (c *Config) With(fields ...zap.Field) *Config {
	c.l = c.l.With(fields...)
	return c
}

// HTTPRequest adds the request fields of [WithHTTPRequest].
func (c *Config) HTTPRequest(r *http.Request) *Config {
	c.l = WithHTTPRequest(c.l, r)
	return c
}
