package telemetry

import (
	"github.com/foomo/keel/markdown"
	"github.com/prometheus/client_golang/prometheus"
)

// Readme returns a markdown table of all metrics gathered from the default
// Prometheus gatherer, or an empty string if there are none.
//
// Deprecated: Readme has no replacement.
func Readme() string {
	md := markdown.Markdown{}

	var rows [][]string

	if gatherer, err := prometheus.DefaultGatherer.Gather(); err == nil {
		for _, value := range gatherer {
			rows = append(rows, []string{
				markdown.Code(value.GetName()),
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
