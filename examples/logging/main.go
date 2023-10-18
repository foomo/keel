package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/service"
)

type CustomError struct {
	error
}

var (
	ErrCustom   = &CustomError{error: errors.New("custom error")}
	ErrStandard = errors.New("string error")
)

func main() {
	svr := keel.NewServer(
		keel.WithHTTPZapService(true),
	)

	// obtain the logger
	l := svr.Logger()

	// alternatively you can always use
	// l := log.Logger()

	// measure tome time
	fDurationFn := log.FDurationFn()
	time.Sleep(200 * time.Millisecond)
	l.Info("measured some time", fDurationFn())

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.WithError(l, ErrCustom).Error("enhanced logger with custom error")
		log.WithError(l, ErrStandard).Error("enhanced logger with standard error")

		log.WithHTTPRequest(l, r).Info("handled request")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}
