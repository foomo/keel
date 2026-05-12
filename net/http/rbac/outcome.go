package rbac

// Outcome is the terminal classification of a request. It drives
// both the HTTP response (allow → pass-through, deny → 403,
// unauthenticated → 401) and the log label.
type Outcome string

const (
	OutcomeAllow           Outcome = "allow"
	OutcomeDeny            Outcome = "deny"
	OutcomeUnauthenticated Outcome = "unauthenticated"
	OutcomeNoRuleAllow     Outcome = "no_rule_allow"
	OutcomeNoRuleDeny      Outcome = "no_rule_deny"
)
