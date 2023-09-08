package metrics

import (
	"github.com/foomo/keel/markdown"
	"github.com/foomo/keel/telemetry/nonrecording"
	"github.com/prometheus/client_golang/prometheus"
)

func Readme() string {
	md := markdown.Markdown{}
	values := nonrecording.Metrics()

	gatherer, _ := prometheus.DefaultRegisterer.(*prometheus.Registry).Gather()
	for _, value := range gatherer {
		values = append(values, nonrecording.Metric{
			Name: value.GetName(),
			Type: value.GetType().String(),
			Help: value.GetHelp(),
		})
	}

	rows := make([][]string, 0, len(values))
	for _, value := range values {
		rows = append(rows, []string{
			markdown.Code(value.Name),
			value.Type,
			value.Help,
		})
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
