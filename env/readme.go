package env

import (
	"fmt"

	"github.com/foomo/keel/markdown"
)

func Readme() string {
	var rows [][]string
	md := &markdown.Markdown{}

	{
		defaults.Range(func(key, fallback any) bool {
			if k, ok := key.(string); ok {
				rows = append(rows, []string{
					markdown.Code(k),
					markdown.Code(TypeOf(k)),
					"",
					markdown.Code(fmt.Sprintf("%v", fallback)),
				})
			}
			return true
		})

		requiredKeys.Range(func(key, fallback any) bool {
			if k, ok := key.(string); ok {
				rows = append(rows, []string{
					markdown.Code(k),
					markdown.Code(TypeOf(k)),
					markdown.Code("true"),
					"",
				})
			}
			return true
		})
	}

	if len(rows) > 0 {
		md.Println("### Env")
		md.Println("")
		md.Println("List of all accessed environment variables.")
		md.Println("")
		md.Table([]string{"Key", "Type", "Required", "Default"}, rows)
		md.Println("")
	}

	return md.String()
}
