package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/env"
)

const (
	ModeDev  = "dev"
	ModeProd = "prod"
)

// logger holds the global logger
var logger *zap.Logger

func init() {
	var config zap.Config
	switch env.Get("LOG", ModeProd) {
	case ModeDev:
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
	default:
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	config.DisableStacktrace = env.GetBool("LOG_DISABLE_STACKTRACE", true)
	config.DisableCaller = env.GetBool("LOG_DISABLE_CALLER", true)
	if l, err := config.Build(); err != nil {
		panic(err)
	} else {
		logger = l
	}
}

// Logger return the logger instance
func Logger() *zap.Logger {
	return logger
}

func Sync() error {
	return logger.Sync()
}

func MustSync() {
	if err := logger.Sync(); err != nil {
		fmt.Println(err)
	}
}

// Must logs a fatal error if given
func Must(l *zap.Logger, err error, msg string) {
	if err != nil {
		if l == nil {
			l = Logger()
		}
		l.Fatal(msg, FError(err), FStackSkip(1))
	}
}

/*
const ()

func LoggerForRequest(l *zap.Logger, r *http.Request) *zap.Logger {
	return l.With(zap.String("http.request", r.URL.String()))
}

func LoggerForGOTSRPCRequest(l *zap.Logger, r *http.Request) *zap.Logger {
	return l.With(zap.String("http.request", r.URL.String()))
}

func LoggerForServiceRest(l *zap.Logger, name string) *zap.Logger {
	return l.With(zap.String("service-rest", name))
}

func LoggerForServiceGOTSRPC(l *zap.Logger, name string) *zap.Logger {
	return l.With(zap.String("service-gotsrpc", name))
}

// _main is the main
func _main() {
	l := LoggerForSquadronService(GetLogger(), "site", "gateway")
	sA := newService(LoggerForService(l, "service-a"))
	sB := newService(LoggerForService(l, "service-b"))
	go sA.serve()
	go sB.serve()
}

type restService struct {
	l *zap.Logger
}

func (s *restService) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	// in a rest service
	requestLogger := LoggerForRequest(s.l, r)
	requestLogger.Info("")
}

func (s *restService) serve() {}

type gotsrpcService struct {
	l *zap.Logger
}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}
}

func LogStuffFromContext(ctx context.Context) (l *zap.Logger, span struct{}) {
	return
}

func (s *gotsrpcService) DoRPCThing(w http.ResponseWriter, r *http.Request, foo string) (baz bool, err error) {
	l, span := LogStuffFromContext(r.Context())
	l.Info("served")
	return
}
*/
