package rbac

// Rule declares allow/deny role lists for a single path pattern.
//
// Path is either an exact request path (e.g. "/api/foo/GetBar") or a
// prefix ending in "*" (e.g. "/api/foo/*"). The "*" matches the empty
// string and any suffix: "/api/foo/*" matches both "/api/foo" and
// "/api/foo/anything".
//
// Semantics applied to (AllowRoles, DenyRoles) against the request's
// roles:
//
//	pass iff (len(AllowRoles) == 0 || roles ∩ AllowRoles ≠ ∅)
//	         && roles ∩ DenyRoles == ∅
//
// An empty AllowRoles list means "no positive constraint" — useful for
// deny-only rules. An empty DenyRoles list means "no explicit
// exclusions". Both empty means "this path is gated by having a
// matching rule at all", which is only meaningful when DefaultPolicy
// is PolicyDeny.
type Rule struct {
	// Exact request path (e.g. "/api/foo/GetBar") or a prefix ending in
	// '*' (e.g. "/api/foo/*"). The '*' suffix matches the empty string and any suffix.
	Path string `yaml:"path" mapstructure:"path" jsonschema:"required,pattern=^/,minLength=1"`
	// Roles granted access. Empty or omitted means "no positive constraint" — useful for deny-only rules.
	AllowRoles []string `yaml:"allowRoles" mapstructure:"allowRoles" jsonschema:"uniqueItems,minLength=1"`
	// Roles denied access. Evaluated before allowRoles; a match here rejects regardless of allowRoles.
	DenyRoles []string `yaml:"denyRoles"  mapstructure:"denyRoles" jsonschema:"uniqueItems,minLength=1"`
}
