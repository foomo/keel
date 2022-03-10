package keel

type ProbeType string

const probesServiceName = "probes"

const (
	ProbeTypeAll        ProbeType = "all"
	ProbeTypeStartup    ProbeType = "startup"
	ProbeTypeReadiness  ProbeType = "readiness"
	ProbeTypeLiveliness ProbeType = "liveliness"
)

func (t ProbeType) String() string {
	return string(t)
}
