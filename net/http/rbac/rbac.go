package rbac

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
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
	exact         map[string]CompiledRBACRule
	prefix        []CompiledRBACRule
}

// CompiledRBACRule is an RBACRule prepared for matching. For prefix
// rules the trailing "*" is stripped once at compile-time so the hot
// path only does prefix comparisons.
//
// Raw is the source rule as declared in configuration; the prefix
// internals stay unexported because they are matcher implementation
// detail callers should not depend on.
type CompiledRBACRule struct {
	Raw      RBACRule
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
		exact:         map[string]CompiledRBACRule{},
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
			m.prefix = append(m.prefix, CompiledRBACRule{Raw: r, prefix: prefix, isPrefix: true})

			continue
		}

		if _, dup := m.exact[r.Path]; dup {
			return nil, fmt.Errorf("rbac: duplicate rule for path %q", r.Path)
		}

		m.exact[r.Path] = CompiledRBACRule{Raw: r}
	}

	// Sort prefixes longest-first so match() can return on the first
	// hit and still be guaranteed to have found the most-specific entry.
	sort.SliceStable(m.prefix, func(i, j int) bool {
		return len(m.prefix[i].prefix) > len(m.prefix[j].prefix)
	})

	return m, nil
}

// RBACOutcome is the terminal classification of a request. It drives
// both the HTTP response (allow → pass-through, deny → 403,
// unauthenticated → 401) and the log label.
type RBACOutcome string

const (
	RBACOutcomeAllow           RBACOutcome = "allow"
	RBACOutcomeDeny            RBACOutcome = "deny"
	RBACOutcomeUnauthenticated RBACOutcome = "unauthenticated"
	RBACOutcomeNoRuleAllow     RBACOutcome = "no_rule_allow"
	RBACOutcomeNoRuleDeny      RBACOutcome = "no_rule_deny"
)

// RbacDecision is the result of evaluating a single request against an
// RBACMatcher. Rule is nil when no rule matched (the default policy
// applied); Roles and Authed are the values returned by the extractor.
type RbacDecision struct {
	Outcome RBACOutcome
	Rule    *CompiledRBACRule
	Roles   []string
	Authed  bool
}

// Evaluate classifies the request against the compiled rule set. The
// 401-vs-403 split is driven by the extractor's authenticated flag.
func (m *RBACMatcher) Evaluate(extract RBACRolesExtractor, r *http.Request) RbacDecision {
	roles, authed := extract(r)

	rule := m.match(r.URL.Path)
	if rule == nil {
		if m.defaultPolicy == RBACPolicyAllow {
			return RbacDecision{Outcome: RBACOutcomeNoRuleAllow, Roles: roles, Authed: authed}
		}

		if !authed {
			return RbacDecision{Outcome: RBACOutcomeUnauthenticated, Roles: roles, Authed: authed}
		}

		return RbacDecision{Outcome: RBACOutcomeNoRuleDeny, Roles: roles, Authed: authed}
	}

	if rbacAllowed(rule.Raw, roles) {
		return RbacDecision{Outcome: RBACOutcomeAllow, Rule: rule, Roles: roles, Authed: authed}
	}

	if !authed {
		return RbacDecision{Outcome: RBACOutcomeUnauthenticated, Rule: rule, Roles: roles, Authed: authed}
	}

	return RbacDecision{Outcome: RBACOutcomeDeny, Rule: rule, Roles: roles, Authed: authed}
}

// match returns the most-specific rule for path: exact match wins, else
// the longest matching prefix wins. Returns nil when no rule applies.
//
// "/api/foo/*" (compiled prefix "/api/foo/") matches both "/api/foo"
// and "/api/foo/anything" — the trailing slash is treated as optional
// for the bare base path.
func (m *RBACMatcher) match(path string) *CompiledRBACRule {
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
