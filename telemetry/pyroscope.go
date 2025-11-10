package telemetry

import (
	"strings"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/attribute"
)

func PyroscopeLabels(kv ...attribute.KeyValue) pyroscope.LabelSet {
	var labels []string

	for _, value := range kv {
		if value.Valid() {
			labels = append(labels, strings.ReplaceAll(string(value.Key), ".", "_"), value.Value.AsString())
		}
	}

	return pyroscope.Labels(labels...)
}
