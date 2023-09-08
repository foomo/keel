package service

import (
	"net/http"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/markdown"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

const (
	DefaultHTTPDocsName = "docs"
	DefaultHTTPDocsAddr = "localhost:9001"
	DefaultHTTPDocsPath = "/docs"
)

func NewHTTPDocs(l *zap.Logger, name, addr, path string, documenters map[string]interfaces.Documenter) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		l.Info("ping  ")
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "text/markdown")
			md := &markdown.Markdown{}
			for name, documenter := range documenters {
				md.Printf("# %s", name)
				md.Println("")
				md.Print(documenter.Docs())
			}
			_, _ = w.Write([]byte(md.String()))
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	return NewHTTP(l, name, addr, handler)
}

func NewDefaultHTTPDocs(documenter map[string]interfaces.Documenter) *HTTP {
	return NewHTTPDocs(
		log.Logger(),
		DefaultHTTPDocsName,
		DefaultHTTPDocsAddr,
		DefaultHTTPDocsPath,
		documenter,
	)
}
