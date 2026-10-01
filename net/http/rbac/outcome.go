package rbac

// Outcome is the terminal classification of a request. It drives
// both the HTTP response (allow → pass-through, deny → 403,
// unauthenticated → 401) and the log label.
type Outcome string

const (
	// OutcomeAllow means a rule matched and the caller's roles satisfy it.
	OutcomeAllow Outcome = "allow"
	// OutcomeDeny means a rule matched, the caller is authenticated and its
	// roles do not satisfy the rule (403).
	OutcomeDeny Outcome = "deny"
	// OutcomeUnauthenticated means the request was denied, by a rule or by
	// [PolicyDeny], and the caller is not authenticated (401).
	OutcomeUnauthenticated Outcome = "unauthenticated"
	// OutcomeNoRuleAllow means no rule matched and [PolicyAllow] let the
	// request through.
	OutcomeNoRuleAllow Outcome = "no_rule_allow"
	// OutcomeNoRuleDeny means no rule matched, the caller is authenticated and
	// [PolicyDeny] rejected the request (403).
	OutcomeNoRuleDeny Outcome = "no_rule_deny"
)
