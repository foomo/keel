package rbac

// Policy is the fallback action applied when no Rule matches the
// request path. Encoded as a string so YAML / JSON configuration decodes
// directly without a custom unmarshaler.
type Policy string

const (
	// PolicyAllow lets unmatched requests through to the next handler.
	PolicyAllow Policy = "allow"
	// PolicyDeny rejects unmatched requests with 401 or 403 depending
	// on whether the caller is authenticated.
	PolicyDeny Policy = "deny"
)
