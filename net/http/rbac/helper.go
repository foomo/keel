package rbac

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfigFromFile reads an Config from the YAML file at path.
//
// Expected file shape:
//
//	defaultPolicy: allow      # or "deny"
//	rules:
//	  - path: "/api/foo/*"
//	    allow: [admin]
//	    deny:  [visitor]
//
// LoadConfigFromFile only decodes the YAML; semantic errors (empty
// paths, duplicate paths, an unknown DefaultPolicy) surface from
// NewMatcher. Keeping load and validation separate lets callers
// inspect the raw decoded shape before committing to a matcher.
func LoadConfigFromFile(path string) (Config, error) {
	var cfg Config

	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("rbac: read %q: %w", path, err)
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("rbac: decode %q: %w", path, err)
	}

	return cfg, nil
}

// rbacAllowed applies a rule's AllowRoles/DenyRoles lists to the roles.
//
//	pass iff (len(AllowRoles) == 0 || roles ∩ AllowRoles ≠ ∅)
//	         && roles ∩ DenyRoles == ∅
func rbacAllowed(r Rule, roles []string) bool {
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
