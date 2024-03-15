package keelmongo

import (
	"strings"

	"github.com/foomo/keel/markdown"
)

var (
	dbs     = map[string][]string{}
	indices = map[string]map[string][]string{}
)

func Readme() string {
	var rows [][]string
	md := &markdown.Markdown{}

	for db, collections := range dbs {
		for _, collection := range collections {
			var i string
			if v, ok := indices[db][collection]; ok {
				i += strings.Join(v, "`, `")
			}
			rows = append(rows, []string{
				markdown.Code(db),
				markdown.Code(collection),
				markdown.Code(i),
			})
		}
	}

	if len(rows) > 0 {
		md.Println("### Mongo")
		md.Println("")
		md.Println("List of all used mongo collections including the configured indices options.")
		md.Println("")
		md.Table([]string{"Database", "Collection", "Indices"}, rows)
		md.Println("")
	}

	return md.String()
}
