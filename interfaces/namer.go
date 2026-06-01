package interfaces

// Namer is implemented by any value that exposes a stable, human-readable
// identifier through a Name method. It is typically used for logging,
// metric labels and registry lookups.
type Namer interface {
	Name() string
}

// IsNamer reports whether v implements [Namer] and returns the asserted
// value. The boolean is false when v does not implement [Namer].
func IsNamer(v any) (Namer, bool) {
	return Is[Namer](v)
}
