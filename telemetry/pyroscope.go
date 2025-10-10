package telemetry

import (
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/attribute"
)

func PyroscopeLabels(kv ...attribute.KeyValue) pyroscope.LabelSet {
	var labels []string

	for _, value := range kv {
		if value.Valid() {
			labels = append(labels, string(value.Key), value.Value.AsString())
		}
	}

	return pyroscope.Labels(labels...)
}
