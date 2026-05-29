package interfaces

// Readmer is implemented by any value that can describe itself through a
// Readme method returning human-readable documentation, typically rendered
// on a diagnostic HTTP endpoint.
type Readmer interface {
	Readme() string
}

// ReadmeHandler adapts a func() string into a value satisfying [Readmer].
// Use [ReadmeFunc] to construct one.
type ReadmeHandler struct {
	Value func() string
}

// Readme calls the wrapped func and returns its result, so ReadmeHandler
// satisfies [Readmer].
func (r ReadmeHandler) Readme() string {
	return r.Value()
}

// ReadmeFunc returns a [ReadmeHandler] that calls v whenever Readme is
// invoked.
func ReadmeFunc(v func() string) ReadmeHandler {
	return ReadmeHandler{
		Value: v,
	}
}
