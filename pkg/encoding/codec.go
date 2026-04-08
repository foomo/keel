package encoding

// Codec encodes T to []byte and decodes []byte back to T.
type Codec[T any] interface {
	Encode(v T) ([]byte, error)
	Decode(b []byte, v *T) error
}
