package config

import (
	"fmt"

	"github.com/foomo/keel/markdown"
)

func Readme() string {
	var configRows [][]string
	var remoteRows [][]string
	c := Config()
	md := &markdown.Markdown{}

	{
		keys := c.AllKeys()
		for _, key := range keys {
			var fallback interface{}
			if v, ok := defaults[key]; ok {
				fallback = v
			}
			configRows = append(configRows, []string{
				markdown.Code(key),
				markdown.Code(TypeOf(key)),
				"",
				markdown.Code(fmt.Sprintf("%v", fallback)),
			})
		}

		for _, key := range requiredKeys {
			configRows = append(configRows, []string{
				markdown.Code(key),
				markdown.Code(TypeOf(key)),
				markdown.Code("true"),
				"",
			})
		}
	}

	{
		for _, remote := range remotes {
			remoteRows = append(remoteRows, []string{
				markdown.Code(remote.provider),
				markdown.Code(remote.path),
			})
		}
	}

	if len(configRows) > 0 || len(remoteRows) > 0 {
		md.Println("### Config")
		md.Println("")
	}

	if len(configRows) > 0 {
		md.Println("List of all registered config variabled with their defaults.")
		md.Println("")
		md.Table([]string{"Key", "Type", "Required", "Default"}, configRows)
		md.Println("")
	}

	if len(remoteRows) > 0 {
		md.Println("#### Remotes")
		md.Println("")
		md.Println("List of remote config providers that are being watched.")
		md.Println("")
		md.Table([]string{"Provider", "Path"}, remoteRows)
		md.Println("")
	}

	return md.String()
}
