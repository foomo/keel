package middleware

import (
	"net/http"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/rbac"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RBAC middleware: a generic, path-based HTTP RBAC gate.
//
// The middleware is transport-agnostic: only r.URL.Path and the
// caller-supplied roles drive the decision. It has no knowledge of any
// specific identity model — callers provide both the roles extractor
// and the configuration.
//
// Typical wiring:
//
//	cfg, err := middleware.LoadRBACConfigFromFile("/etc/rbac.yaml")
//	keellog.Must(l, err, "load rbac config")
//
//	matcher, err := middleware.NewRBACMatcher(cfg)
//	keellog.Must(l, err, "compile rbac config")
//
//	mw := middleware.RBAC(matcher, extractRoles)
//
// m and extract must be non-nil; they encode the decision contract and
// the middleware does not validate them per request.
//
// Rule semantics: a request passes iff it hits no DenyRoles entry AND
// either the rule has no AllowRoles entries or roles intersect the
// AllowRoles set. Path matching picks the most specific rule (exact
// wins over longest matching prefix).
//
// Denied requests log a structured zap warning ("rbac denied request")
// at WarnLevel with path, outcome, authenticated flag, roles, and
// matched rule path; allowed requests are not logged. The 401/403
// split follows the extractor's authenticated flag: unauthenticated →
// 401, authenticated-but-disallowed → 403.
func RBAC(m *rbac.RBACMatcher, extract rbac.RBACRolesExtractor) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("RBAC")
			}

			d := m.Evaluate(extract, r)

			// Fast path: allowed requests skip logging entirely — they
			// are the common case and would dwarf denies under traffic.
			if d.Outcome == rbac.RBACOutcomeAllow || d.Outcome == rbac.RBACOutcomeNoRuleAllow {
				next.ServeHTTP(w, r)
				return
			}

			if l != nil {
				fields := []zap.Field{
					zap.String("path", r.URL.Path),
					zap.String("outcome", string(d.Outcome)),
					zap.Bool("authenticated", d.Authed),
					zap.Strings("roles", d.Roles),
				}
				if d.Rule != nil {
					fields = append(fields, zap.String("rule_path", d.Rule.Raw.Path))
				}

				l.Warn("rbac denied request", fields...)
			}

			status := http.StatusForbidden
			if d.Outcome == rbac.RBACOutcomeUnauthenticated {
				status = http.StatusUnauthorized
			}

			http.Error(w, http.StatusText(status), status)
		})
	}
}
