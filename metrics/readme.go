package metrics

import (
	"github.com/foomo/keel/markdown"
	"github.com/prometheus/client_golang/prometheus"
)

func Readme() string {
	md := markdown.Markdown{}
	var rows [][]string

	if gatherer, err := prometheus.DefaultGatherer.Gather(); err == nil {
		for _, value := range gatherer {
			rows = append(rows, []string{
				value.GetName(),
				value.GetType().String(),
				value.GetHelp(),
			})
		}
	}

	if len(rows) > 0 {
		md.Println("### Metrics")
		md.Println("")
		md.Println("List of all registered metrics than are being exposed.")
		md.Println("")
		md.Table([]string{"Name", "Type", "Description"}, rows)
		md.Println("")
	}

	return md.String()
}
