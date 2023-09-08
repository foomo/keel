package service

import (
	"net/http"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/markdown"
	"go.uber.org/zap"
)

const (
	DefaultHTTPDocsName = "docs"
	DefaultHTTPDocsAddr = "localhost:9001"
	DefaultHTTPDocsPath = "/docs"
)

func NewHTTPDocs(l *zap.Logger, name, addr, path string, documenters map[string]interfaces.Documenter) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "text/markdown")
			md := &markdown.Markdown{}
			for name, documenter := range documenters {
				md.Printf("## %s", name)
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

func NewDefaultHTTPDocs(l *zap.Logger, documenter map[string]interfaces.Documenter) *HTTP {
	return NewHTTPDocs(
		l,
		DefaultHTTPDocsName,
		DefaultHTTPDocsAddr,
		DefaultHTTPDocsPath,
		documenter,
	)
}
