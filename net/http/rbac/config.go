package rbac

// Config is the file-shape configuration for the middleware.
//
// DefaultPolicy applies when no rule matches the request path. It must
// be PolicyAllow or PolicyDeny — the empty string is rejected by
// NewMatcher so callers make an explicit choice.
type Config struct {
	// Action applied when no rule matches the request path
	DefaultPolicy Policy `yaml:"defaultPolicy" mapstructure:"defaultPolicy" jsonschema:"required,enum=allow,enum=deny"`
	// Path-keyed allow/deny rules. Path matching prefers exact paths, then the longest matching prefix
	Rules []Rule `yaml:"rules" mapstructure:"rules"`
}
