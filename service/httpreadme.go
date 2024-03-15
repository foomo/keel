package service

import (
	"net/http"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/markdown"
	"go.uber.org/zap"
)

var (
	DefaultHTTPReadmeName = "readme"
	DefaultHTTPReadmeAddr = "localhost:9001"
	DefaultHTTPReadmePath = "/readme"
)

func NewHTTPReadme(l *zap.Logger, name, addr, path string, readmers func() []interfaces.Readmer) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Add("Content-Type", "text/markdown")
			w.WriteHeader(http.StatusOK)
			md := &markdown.Markdown{}
			for _, readmer := range readmers() {
				md.Print(readmer.Readme())
			}
			_, _ = w.Write([]byte(md.String()))
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	return NewHTTP(l, name, addr, handler)
}

func NewDefaultHTTPReadme(l *zap.Logger, readmers func() []interfaces.Readmer) *HTTP {
	return NewHTTPReadme(
		l,
		DefaultHTTPReadmeName,
		DefaultHTTPReadmeAddr,
		DefaultHTTPReadmePath,
		readmers,
	)
}
