// Package rbac provides path-based, role-driven access control rules for
// HTTP requests.
//
// A [Config] declares a default [Policy] and a list of [Rule] values keyed by
// request path. [NewMatcher] validates and compiles it into a [Matcher], which
// classifies each request into an [Outcome] using the roles returned by a
// [RolesExtractor]. The [github.com/foomo/keel/net/http/middleware.RBAC]
// middleware turns these outcomes into pass-through, 401 or 403 responses.
//
// # Path matching
//
// An exact path rule takes precedence over prefix rules (paths ending in
// "*"), and among prefix rules the longest matching prefix wins. When no rule
// matches, the default policy applies.
//
// # Configuration
//
// Configurations are typically loaded from YAML with [LoadConfigFromFile]:
//
//	defaultPolicy: deny
//	rules:
//	  - path: "/api/admin/*"
//	    allowRoles: [admin]
//	  - path: "/api/public/*"
//
// The file rbac.schema.json in this package is a JSON schema for editor
// validation of such files.
package rbac
