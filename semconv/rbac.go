package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	// KeelRBACOutcomeKey is the key for keel.rbac.outcome.
	KeelRBACOutcomeKey = attribute.Key("keel.rbac.outcome")
	// KeelRBACAuthenticatedKey is the key for keel.rbac.authenticated.
	KeelRBACAuthenticatedKey = attribute.Key("keel.rbac.authenticated")
	// KeelRBACRolesKey is the key for keel.rbac.roles.
	KeelRBACRolesKey = attribute.Key("keel.rbac.roles")
	// KeelRBACRulePathKey is the key for keel.rbac.rule.path.
	KeelRBACRulePathKey = attribute.Key("keel.rbac.rule.path")
)

// KeelRBACOutcome returns a new attribute.KeyValue for keel.rbac.outcome.
func KeelRBACOutcome(v string) attribute.KeyValue {
	return KeelRBACOutcomeKey.String(v)
}

// KeelRBACAuthenticated returns a new attribute.KeyValue for keel.rbac.authenticated.
func KeelRBACAuthenticated(v bool) attribute.KeyValue {
	return KeelRBACAuthenticatedKey.Bool(v)
}

// KeelRBACRoles returns a new attribute.KeyValue for keel.rbac.roles.
func KeelRBACRoles(v []string) attribute.KeyValue {
	return KeelRBACRolesKey.StringSlice(v)
}

// KeelRBACRulePath returns a new attribute.KeyValue for keel.rbac.rule.path.
func KeelRBACRulePath(v string) attribute.KeyValue {
	return KeelRBACRulePathKey.String(v)
}
