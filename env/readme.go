package env

import (
	"fmt"

	"github.com/foomo/keel/markdown"
)

func Readme() string {
	var rows [][]string
	md := &markdown.Markdown{}

	{
		for key, fallback := range defaults {
			rows = append(rows, []string{
				markdown.Code(key),
				markdown.Code(TypeOf(key)),
				"",
				markdown.Code(fmt.Sprintf("%v", fallback)),
			})
		}

		for _, key := range requiredKeys {
			rows = append(rows, []string{
				markdown.Code(key),
				markdown.Code(TypeOf(key)),
				markdown.Code("true"),
				"",
			})
		}
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
