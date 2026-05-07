package middleware

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	keelhttp "github.com/foomo/keel/net/http"
)

// RBACPolicy is the fallback action applied when no RBACRule matches the
// request path. Encoded as a string so YAML / JSON configuration decodes
// directly without a custom unmarshaler.
type RBACPolicy string

const (
	// RBACPolicyAllow lets unmatched requests through to the next handler.
	RBACPolicyAllow RBACPolicy = "allow"
	// RBACPolicyDeny rejects unmatched requests with 401 or 403 depending
	// on whether the caller is authenticated.
	RBACPolicyDeny RBACPolicy = "deny"
)

// RBACRule declares allow/deny role lists for a single path pattern.
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
// is RBACPolicyDeny.
type RBACRule struct {
	Path       string   `yaml:"path"       mapstructure:"path"`
	AllowRoles []string `yaml:"allowRoles" mapstructure:"allowRoles"`
	DenyRoles  []string `yaml:"denyRoles"  mapstructure:"denyRoles"`
}

// RBACConfig is the file-shape configuration for the middleware.
//
// DefaultPolicy applies when no rule matches the request path. It must
// be RBACPolicyAllow or RBACPolicyDeny — the empty string is rejected by
// NewRBACMatcher so callers make an explicit choice.
type RBACConfig struct {
	DefaultPolicy RBACPolicy `yaml:"defaultPolicy" mapstructure:"defaultPolicy"`
	Rules         []RBACRule `yaml:"rules"         mapstructure:"rules"`
}

// RBACRolesExtractor pulls the set of roles attached to a request,
// along with an authenticated flag that drives the 401-vs-403
// distinction on denial.
//
// "Roles" is a deliberately broad label: any string the caller wants to
// match against AllowRoles / DenyRoles fits — IAM roles, group
// memberships, OAuth scopes, custom claim values, and so on. The RBAC
// package name commits to role-shaped vocabulary; the extractor stays
// agnostic about where the values come from.
//
// Contract:
//   - roles may be nil or empty for an authenticated caller with no
//     matching roles; the middleware treats len(roles) == 0 as
//     "no positive matches" rather than as unauthenticated.
//   - authenticated reflects whether the request carries a trusted
//     identity at all. An extractor that cannot tell the difference
//     should return true whenever it returns any roles, and false
//     only on a confirmed absence of identity.
//   - the extractor must not mutate the request.
type RBACRolesExtractor func(r *http.Request) (roles []string, authenticated bool)

// RBACMatcher is a validated, compiled RBACConfig ready to drive the
// middleware. Build one via NewRBACMatcher and pass it to RBAC.
//
// RBACMatcher is safe for concurrent use: it is read-only after
// construction.
type RBACMatcher struct {
	defaultPolicy RBACPolicy
	exact         map[string]compiledRBACRule
	prefix        []compiledRBACRule
}

// compiledRBACRule is an RBACRule prepared for matching. For prefix
// rules the trailing "*" is stripped once at compile-time so the hot
// path only does prefix comparisons.
type compiledRBACRule struct {
	raw      RBACRule
	prefix   string // populated iff isPrefix
	isPrefix bool
}

// NewRBACMatcher validates cfg and compiles it into a matcher.
//
// Returns an error when:
//   - cfg.DefaultPolicy is neither RBACPolicyAllow nor RBACPolicyDeny
//     (including the zero-value empty string).
//   - any RBACRule has an empty Path.
//   - two rules share the same exact path or the same prefix.
//
// All validation errors surface here so callers can fail fast at
// service startup rather than at request time.
func NewRBACMatcher(cfg RBACConfig) (*RBACMatcher, error) {
	switch cfg.DefaultPolicy {
	case RBACPolicyAllow, RBACPolicyDeny:
	default:
		return nil, fmt.Errorf("rbac: invalid DefaultPolicy %q (want %q or %q)", cfg.DefaultPolicy, RBACPolicyAllow, RBACPolicyDeny)
	}

	m := &RBACMatcher{
		defaultPolicy: cfg.DefaultPolicy,
		exact:         map[string]compiledRBACRule{},
	}
	// Track prefixes separately from exact paths so "/api/foo" (exact)
	// and "/api/foo/*" (prefix "/api/foo/") don't collide as duplicates.
	prefixSeen := map[string]struct{}{}

	for i, r := range cfg.Rules {
		if r.Path == "" {
			return nil, fmt.Errorf("rbac: rule #%d has empty path", i)
		}

		if before, ok := strings.CutSuffix(r.Path, "*"); ok {
			prefix := before
			if _, dup := prefixSeen[prefix]; dup {
				return nil, fmt.Errorf("rbac: duplicate rule for path %q", r.Path)
			}

			prefixSeen[prefix] = struct{}{}
			m.prefix = append(m.prefix, compiledRBACRule{raw: r, prefix: prefix, isPrefix: true})

			continue
		}

		if _, dup := m.exact[r.Path]; dup {
			return nil, fmt.Errorf("rbac: duplicate rule for path %q", r.Path)
		}

		m.exact[r.Path] = compiledRBACRule{raw: r}
	}

	// Sort prefixes longest-first so match() can return on the first
	// hit and still be guaranteed to have found the most-specific entry.
	sort.SliceStable(m.prefix, func(i, j int) bool {
		return len(m.prefix[i].prefix) > len(m.prefix[j].prefix)
	})

	return m, nil
}

// match returns the most-specific rule for path: exact match wins, else
// the longest matching prefix wins. Returns nil when no rule applies.
//
// "/api/foo/*" (compiled prefix "/api/foo/") matches both "/api/foo"
// and "/api/foo/anything" — the trailing slash is treated as optional
// for the bare base path.
func (m *RBACMatcher) match(path string) *compiledRBACRule {
	if r, ok := m.exact[path]; ok {
		return &r
	}

	for i := range m.prefix {
		p := m.prefix[i].prefix
		if strings.HasPrefix(path, p) ||
			(strings.HasSuffix(p, "/") && path == strings.TrimRight(p, "/")) {
			return &m.prefix[i]
		}
	}

	return nil
}

// rbacOutcome is the terminal classification of a request. It drives
// both the HTTP response (allow → pass-through, deny → 403,
// unauthenticated → 401) and the log label.
type rbacOutcome string

const (
	rbacOutcomeAllow           rbacOutcome = "allow"
	rbacOutcomeDeny            rbacOutcome = "deny"
	rbacOutcomeUnauthenticated rbacOutcome = "unauthenticated"
	rbacOutcomeNoRuleAllow     rbacOutcome = "no_rule_allow"
	rbacOutcomeNoRuleDeny      rbacOutcome = "no_rule_deny"
)

type rbacDecision struct {
	outcome rbacOutcome
	rule    *compiledRBACRule
	roles   []string
	authed  bool
}

// evaluate classifies the request against the compiled rule set. The
// 401-vs-403 split is driven by the extractor's authenticated flag.
func (m *RBACMatcher) evaluate(extract RBACRolesExtractor, r *http.Request) rbacDecision {
	roles, authed := extract(r)

	rule := m.match(r.URL.Path)
	if rule == nil {
		if m.defaultPolicy == RBACPolicyAllow {
			return rbacDecision{outcome: rbacOutcomeNoRuleAllow, roles: roles, authed: authed}
		}

		if !authed {
			return rbacDecision{outcome: rbacOutcomeUnauthenticated, roles: roles, authed: authed}
		}

		return rbacDecision{outcome: rbacOutcomeNoRuleDeny, roles: roles, authed: authed}
	}

	if rbacAllowed(rule.raw, roles) {
		return rbacDecision{outcome: rbacOutcomeAllow, rule: rule, roles: roles, authed: authed}
	}

	if !authed {
		return rbacDecision{outcome: rbacOutcomeUnauthenticated, rule: rule, roles: roles, authed: authed}
	}

	return rbacDecision{outcome: rbacOutcomeDeny, rule: rule, roles: roles, authed: authed}
}

// rbacAllowed applies a rule's AllowRoles/DenyRoles lists to the roles.
//
//	pass iff (len(AllowRoles) == 0 || roles ∩ AllowRoles ≠ ∅)
//	         && roles ∩ DenyRoles == ∅
func rbacAllowed(r RBACRule, roles []string) bool {
	if rbacIntersects(r.DenyRoles, roles) {
		return false
	}

	if len(r.AllowRoles) == 0 {
		return true
	}

	return rbacIntersects(r.AllowRoles, roles)
}

func rbacIntersects(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}

	return false
}

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
func RBAC(m *RBACMatcher, extract RBACRolesExtractor) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("RBAC")
			}

			d := m.evaluate(extract, r)

			// Fast path: allowed requests skip logging entirely — they
			// are the common case and would dwarf denies under traffic.
			if d.outcome == rbacOutcomeAllow || d.outcome == rbacOutcomeNoRuleAllow {
				next.ServeHTTP(w, r)
				return
			}

			if l != nil {
				fields := []zap.Field{
					zap.String("path", r.URL.Path),
					zap.String("outcome", string(d.outcome)),
					zap.Bool("authenticated", d.authed),
					zap.Strings("roles", d.roles),
				}
				if d.rule != nil {
					fields = append(fields, zap.String("rule_path", d.rule.raw.Path))
				}

				l.Warn("rbac denied request", fields...)
			}

			status := http.StatusForbidden
			if d.outcome == rbacOutcomeUnauthenticated {
				status = http.StatusUnauthorized
			}

			http.Error(w, http.StatusText(status), status)
		})
	}
}

// LoadRBACConfigFromFile reads an RBACConfig from the YAML file at path.
//
// Expected file shape:
//
//	defaultPolicy: allow      # or "deny"
//	rules:
//	  - path: "/api/foo/*"
//	    allow: [admin]
//	    deny:  [visitor]
//
// LoadRBACConfigFromFile only decodes the YAML; semantic errors (empty
// paths, duplicate paths, an unknown DefaultPolicy) surface from
// NewRBACMatcher. Keeping load and validation separate lets callers
// inspect the raw decoded shape before committing to a matcher.
func LoadRBACConfigFromFile(path string) (RBACConfig, error) {
	var cfg RBACConfig

	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("rbac: read %q: %w", path, err)
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("rbac: decode %q: %w", path, err)
	}

	return cfg, nil
}
