package interfaces

// Readmer interface
type Readmer interface {
	Readme() string
}

type ReadmeHandler struct {
	Value func() string
}

func (r ReadmeHandler) Readme() string {
	return r.Value()
}

func ReadmeFunc(v func() string) ReadmeHandler {
	return ReadmeHandler{
		Value: v,
	}
}
