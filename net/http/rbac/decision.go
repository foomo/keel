package rbac

// Decision is the result of evaluating a single request against a
// [Matcher]. Rule is nil when no rule matched (the default policy
// applied); Roles and Authed are the values returned by the extractor.
type Decision struct {
	Outcome Outcome
	Rule    *CompiledRule
	Roles   []string
	Authed  bool
}
