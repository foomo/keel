package telemetry

var (
	// Deprecated: DefaultHistogramBuckets units are selected for metrics in "seconds" unit
	DefaultHistogramBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25, 60, 120, 300, 600}
)
