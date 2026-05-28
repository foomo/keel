package rbac_test

import (
	"testing"

	"github.com/foomo/keel/net/http/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMatcher_invalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         rbac.Config
		errFragment string
	}{
		{
			name:        "missing default policy",
			cfg:         rbac.Config{},
			errFragment: "DefaultPolicy",
		},
		{
			name:        "unknown default policy",
			cfg:         rbac.Config{DefaultPolicy: rbac.Policy("bogus")},
			errFragment: "bogus",
		},
		{
			name: "empty rule path",
			cfg: rbac.Config{
				DefaultPolicy: rbac.PolicyAllow,
				Rules:         []rbac.Rule{{Path: ""}},
			},
			errFragment: "empty path",
		},
		{
			name: "duplicate exact path",
			cfg: rbac.Config{
				DefaultPolicy: rbac.PolicyAllow,
				Rules: []rbac.Rule{
					{Path: "/api/foo"},
					{Path: "/api/foo"},
				},
			},
			errFragment: "duplicate",
		},
		{
			name: "duplicate prefix path",
			cfg: rbac.Config{
				DefaultPolicy: rbac.PolicyAllow,
				Rules: []rbac.Rule{
					{Path: "/api/foo/*"},
					{Path: "/api/foo/*"},
				},
			},
			errFragment: "duplicate",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := rbac.NewMatcher(tc.cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errFragment)
		})
	}
}

func TestNewMatcher_acceptsExactAndPrefixForSameBase(t *testing.T) {
	t.Parallel()

	// "/api/foo" (exact) and "/api/foo/*" (prefix "/api/foo/") are not
	// duplicates — the prefix has a trailing slash that the exact path
	// does not.
	_, err := rbac.NewMatcher(rbac.Config{
		DefaultPolicy: rbac.PolicyAllow,
		Rules: []rbac.Rule{
			{Path: "/api/foo"},
			{Path: "/api/foo/*"},
		},
	})
	require.NoError(t, err)
}
