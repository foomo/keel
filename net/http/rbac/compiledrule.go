package rbac

// CompiledRule is an Rule prepared for matching. For prefix
// rules the trailing "*" is stripped once at compile-time so the hot
// path only does prefix comparisons.
//
// Raw is the source rule as declared in configuration; the prefix
// internals stay unexported because they are matcher implementation
// detail callers should not depend on.
type CompiledRule struct {
	Raw      Rule
	prefix   string // populated iff isPrefix
	isPrefix bool
}
