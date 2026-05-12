package rbac

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Matcher is a validated, compiled Config ready to drive the
// middleware. Build one via NewMatcher and pass it to RBAC.
//
// Matcher is safe for concurrent use: it is read-only after
// construction.
type Matcher struct {
	defaultPolicy Policy
	exact         map[string]CompiledRule
	prefix        []CompiledRule
}

// NewMatcher validates cfg and compiles it into a matcher.
//
// Returns an error when:
//   - cfg.DefaultPolicy is neither PolicyAllow nor PolicyDeny
//     (including the zero-value empty string).
//   - any Rule has an empty Path.
//   - two rules share the same exact path or the same prefix.
//
// All validation errors surface here so callers can fail fast at
// service startup rather than at request time.
func NewMatcher(cfg Config) (*Matcher, error) {
	switch cfg.DefaultPolicy {
	case PolicyAllow, PolicyDeny:
	default:
		return nil, fmt.Errorf("rbac: invalid DefaultPolicy %q (want %q or %q)", cfg.DefaultPolicy, PolicyAllow, PolicyDeny)
	}

	m := &Matcher{
		defaultPolicy: cfg.DefaultPolicy,
		exact:         map[string]CompiledRule{},
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
			m.prefix = append(m.prefix, CompiledRule{Raw: r, prefix: prefix, isPrefix: true})

			continue
		}

		if _, dup := m.exact[r.Path]; dup {
			return nil, fmt.Errorf("rbac: duplicate rule for path %q", r.Path)
		}

		m.exact[r.Path] = CompiledRule{Raw: r}
	}

	// Sort prefixes longest-first so match() can return on the first
	// hit and still be guaranteed to have found the most-specific entry.
	sort.SliceStable(m.prefix, func(i, j int) bool {
		return len(m.prefix[i].prefix) > len(m.prefix[j].prefix)
	})

	return m, nil
}

// Evaluate classifies the request against the compiled rule set. The
// 401-vs-403 split is driven by the extractor's authenticated flag.
func (m *Matcher) Evaluate(extract RolesExtractor, r *http.Request) Decision {
	roles, authed := extract(r)

	rule := m.match(r.URL.Path)
	if rule == nil {
		if m.defaultPolicy == PolicyAllow {
			return Decision{Outcome: OutcomeNoRuleAllow, Roles: roles, Authed: authed}
		}

		if !authed {
			return Decision{Outcome: OutcomeUnauthenticated, Roles: roles, Authed: authed}
		}

		return Decision{Outcome: OutcomeNoRuleDeny, Roles: roles, Authed: authed}
	}

	if rbacAllowed(rule.Raw, roles) {
		return Decision{Outcome: OutcomeAllow, Rule: rule, Roles: roles, Authed: authed}
	}

	if !authed {
		return Decision{Outcome: OutcomeUnauthenticated, Rule: rule, Roles: roles, Authed: authed}
	}

	return Decision{Outcome: OutcomeDeny, Rule: rule, Roles: roles, Authed: authed}
}

// match returns the most-specific rule for path: exact match wins, else
// the longest matching prefix wins. Returns nil when no rule applies.
//
// "/api/foo/*" (compiled prefix "/api/foo/") matches both "/api/foo"
// and "/api/foo/anything" — the trailing slash is treated as optional
// for the bare base path.
func (m *Matcher) match(path string) *CompiledRule {
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
