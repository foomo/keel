package service

import (
	"net/http"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/markdown"
	"go.uber.org/zap"
)

// Default name, address and path of the service returned by
// [NewDefaultHTTPReadme].
var (
	DefaultHTTPReadmeName = "readme"
	DefaultHTTPReadmeAddr = "localhost:9001"
	DefaultHTTPReadmePath = "/readme"
)

// NewHTTPReadme returns an [HTTP] service that responds to GET requests on path
// with the concatenated markdown of the readmers returned by readmers.
//
// Deprecated: Readme support will be removed in a future release.
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

// NewDefaultHTTPReadme returns [NewHTTPReadme] using [DefaultHTTPReadmeName],
// [DefaultHTTPReadmeAddr] and [DefaultHTTPReadmePath].
//
// Deprecated: Readme support will be removed in a future release.
func NewDefaultHTTPReadme(l *zap.Logger, readmers func() []interfaces.Readmer) *HTTP {
	return NewHTTPReadme(
		l,
		DefaultHTTPReadmeName,
		DefaultHTTPReadmeAddr,
		DefaultHTTPReadmePath,
		readmers,
	)
}
