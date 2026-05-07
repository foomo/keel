package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/foomo/keel/net/http/middleware"
)

// staticRBACExtractor returns a fixed (roles, authed) for every
// request — keeps each subtest self-contained without having to wire
// per-test extraction.
func staticRBACExtractor(roles []string, authed bool) middleware.RBACRolesExtractor {
	return func(_ *http.Request) ([]string, bool) { return roles, authed }
}

func newRBACHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func mountRBAC(t *testing.T, cfg middleware.RBACConfig, extract middleware.RBACRolesExtractor) (http.Handler, *observer.ObservedLogs) {
	t.Helper()

	matcher, err := middleware.NewRBACMatcher(cfg)
	require.NoError(t, err)

	core, recorded := observer.New(zap.WarnLevel)
	l := zap.New(core)
	mw := middleware.RBAC(matcher, extract)

	return mw(l, "test", newRBACHandler()), recorded
}

func TestRBAC_outcomes(t *testing.T) {
	t.Parallel()

	cfg := middleware.RBACConfig{
		DefaultPolicy: middleware.RBACPolicyAllow,
		Rules: []middleware.RBACRule{
			{Path: "/admin/*", AllowRoles: []string{"admin"}},
			{Path: "/banned/*", DenyRoles: []string{"banned"}},
			{Path: "/editor/*", AllowRoles: []string{"editor", "admin"}},
		},
	}

	tests := []struct {
		name       string
		path       string
		roles      []string
		authed     bool
		wantStatus int
		wantLog    bool
	}{
		{
			name:       "allow list hit => 200",
			path:       "/admin/users",
			roles:      []string{"admin"},
			authed:     true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "allow list miss, authenticated => 403",
			path:       "/admin/users",
			roles:      []string{"editor"},
			authed:     true,
			wantStatus: http.StatusForbidden,
			wantLog:    true,
		},
		{
			name:       "allow list miss, unauthenticated => 401",
			path:       "/admin/users",
			roles:      nil,
			authed:     false,
			wantStatus: http.StatusUnauthorized,
			wantLog:    true,
		},
		{
			name:       "deny list hit, authenticated => 403",
			path:       "/banned/stuff",
			roles:      []string{"banned"},
			authed:     true,
			wantStatus: http.StatusForbidden,
			wantLog:    true,
		},
		{
			name:       "no matching rule, default allow => 200",
			path:       "/other",
			roles:      nil,
			authed:     false,
			wantStatus: http.StatusOK,
		},
		{
			name:       "editor allow list, user has admin => 200",
			path:       "/editor/page",
			roles:      []string{"admin"},
			authed:     true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "deny list miss with empty allow => 200",
			path:       "/banned/stuff",
			roles:      []string{"editor"},
			authed:     true,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler, logs := mountRBAC(t, cfg, staticRBACExtractor(tc.roles, tc.authed))
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantLog {
				require.Equal(t, 1, logs.Len(), "expected one deny log entry")
				entry := logs.All()[0]
				assert.Equal(t, "rbac denied request", entry.Message)
				assert.Equal(t, zap.WarnLevel, entry.Level)
				assert.Equal(t, tc.path, entry.ContextMap()["path"])
			} else {
				assert.Equal(t, 0, logs.Len(), "expected no log entries")
			}
		})
	}
}

func TestRBAC_defaultDeny(t *testing.T) {
	t.Parallel()

	cfg := middleware.RBACConfig{
		DefaultPolicy: middleware.RBACPolicyDeny,
		Rules: []middleware.RBACRule{
			{Path: "/open/*"},
		},
	}

	t.Run("unmatched path, unauthenticated => 401", func(t *testing.T) {
		t.Parallel()
		handler, _ := mountRBAC(t, cfg, staticRBACExtractor(nil, false))
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/unconfigured", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("unmatched path, authenticated => 403", func(t *testing.T) {
		t.Parallel()
		handler, _ := mountRBAC(t, cfg, staticRBACExtractor([]string{"editor"}, true))
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/unconfigured", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("matched path with no constraints => 200", func(t *testing.T) {
		t.Parallel()
		handler, _ := mountRBAC(t, cfg, staticRBACExtractor(nil, false))
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/open/thing", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// TestRBAC_pathMatching verifies path-precedence rules through the
// public middleware: each rule has a distinct AllowRoles list, and the
// (role, path) combination chosen for each test is one that ONLY
// passes when that specific rule matched. The path that did match is
// also asserted via the deny log entry's "rule_path" field.
func TestRBAC_pathMatching(t *testing.T) {
	t.Parallel()

	cfg := middleware.RBACConfig{
		DefaultPolicy: middleware.RBACPolicyDeny,
		Rules: []middleware.RBACRule{
			{Path: "/api/foo/*", AllowRoles: []string{"foo"}},
			{Path: "/api/foo/AdminOnly", AllowRoles: []string{"admin"}},
			{Path: "/api/foo/bar/*", AllowRoles: []string{"bar"}},
			{Path: "/api/other", AllowRoles: []string{"other"}},
		},
	}

	tests := []struct {
		name     string
		path     string
		roles    []string
		wantPass bool
		wantRule string // expected rule_path in deny log; "" when wantPass
	}{
		{name: "exact wins over prefix", path: "/api/foo/AdminOnly", roles: []string{"admin"}, wantPass: true},
		{name: "exact rejects when only prefix role present", path: "/api/foo/AdminOnly", roles: []string{"foo"}, wantPass: false, wantRule: "/api/foo/AdminOnly"},
		{name: "longest prefix wins", path: "/api/foo/bar/baz", roles: []string{"bar"}, wantPass: true},
		{name: "longest prefix rejects shorter-prefix role", path: "/api/foo/bar/baz", roles: []string{"foo"}, wantPass: false, wantRule: "/api/foo/bar/*"},
		{name: "short prefix matches when no longer one applies", path: "/api/foo/other", roles: []string{"foo"}, wantPass: true},
		{name: "prefix matches bare base path", path: "/api/foo", roles: []string{"foo"}, wantPass: true},
		{name: "exact path", path: "/api/other", roles: []string{"other"}, wantPass: true},
		{name: "no match -> default deny -> 403", path: "/api/unrelated", roles: []string{"foo"}, wantPass: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler, logs := mountRBAC(t, cfg, staticRBACExtractor(tc.roles, true))
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tc.wantPass {
				assert.Equal(t, http.StatusOK, rec.Code)
				return
			}

			assert.Equal(t, http.StatusForbidden, rec.Code)

			if tc.wantRule != "" {
				require.Equal(t, 1, logs.Len(), "expected one deny log entry")
				assert.Equal(t, tc.wantRule, logs.All()[0].ContextMap()["rule_path"])
			}
		})
	}
}

func TestNewRBACMatcher_invalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         middleware.RBACConfig
		errFragment string
	}{
		{
			name:        "missing default policy",
			cfg:         middleware.RBACConfig{},
			errFragment: "DefaultPolicy",
		},
		{
			name:        "unknown default policy",
			cfg:         middleware.RBACConfig{DefaultPolicy: middleware.RBACPolicy("bogus")},
			errFragment: "bogus",
		},
		{
			name: "empty rule path",
			cfg: middleware.RBACConfig{
				DefaultPolicy: middleware.RBACPolicyAllow,
				Rules:         []middleware.RBACRule{{Path: ""}},
			},
			errFragment: "empty path",
		},
		{
			name: "duplicate exact path",
			cfg: middleware.RBACConfig{
				DefaultPolicy: middleware.RBACPolicyAllow,
				Rules: []middleware.RBACRule{
					{Path: "/api/foo"},
					{Path: "/api/foo"},
				},
			},
			errFragment: "duplicate",
		},
		{
			name: "duplicate prefix path",
			cfg: middleware.RBACConfig{
				DefaultPolicy: middleware.RBACPolicyAllow,
				Rules: []middleware.RBACRule{
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

			_, err := middleware.NewRBACMatcher(tc.cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errFragment)
		})
	}
}

func TestNewRBACMatcher_acceptsExactAndPrefixForSameBase(t *testing.T) {
	t.Parallel()

	// "/api/foo" (exact) and "/api/foo/*" (prefix "/api/foo/") are not
	// duplicates — the prefix has a trailing slash that the exact path
	// does not.
	_, err := middleware.NewRBACMatcher(middleware.RBACConfig{
		DefaultPolicy: middleware.RBACPolicyAllow,
		Rules: []middleware.RBACRule{
			{Path: "/api/foo"},
			{Path: "/api/foo/*"},
		},
	})
	require.NoError(t, err)
}

func writeRBACConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "rbac.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	return path
}

func TestLoadRBACConfigFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		path := writeRBACConfig(t, `
defaultPolicy: deny
rules:
  - path: "/api/foo/*"
    allowRoles: [admin, contentEditor]
  - path: "/api/bar/AdminOnly"
    allowRoles: [admin]
    denyRoles:  [visitor]
`)
		cfg, err := middleware.LoadRBACConfigFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, middleware.RBACPolicyDeny, cfg.DefaultPolicy)
		require.Len(t, cfg.Rules, 2)
		assert.Equal(t, "/api/foo/*", cfg.Rules[0].Path)
		assert.ElementsMatch(t, []string{"admin", "contentEditor"}, cfg.Rules[0].AllowRoles)
		assert.Equal(t, "/api/bar/AdminOnly", cfg.Rules[1].Path)
		assert.ElementsMatch(t, []string{"admin"}, cfg.Rules[1].AllowRoles)
		assert.ElementsMatch(t, []string{"visitor"}, cfg.Rules[1].DenyRoles)
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		_, err := middleware.LoadRBACConfigFromFile(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read")
	})

	t.Run("malformed yaml", func(t *testing.T) {
		t.Parallel()

		path := writeRBACConfig(t, `defaultPolicy: allow
rules: [this is not valid yaml`)
		_, err := middleware.LoadRBACConfigFromFile(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode")
	})

	t.Run("empty rules", func(t *testing.T) {
		t.Parallel()

		path := writeRBACConfig(t, "defaultPolicy: allow\nrules: []\n")
		cfg, err := middleware.LoadRBACConfigFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, middleware.RBACPolicyAllow, cfg.DefaultPolicy)
		assert.Empty(t, cfg.Rules)
	})

	// Structural errors (bogus policy, empty path) decode fine —
	// they surface from NewRBACMatcher, not from the loader.
	t.Run("structural errors surface in NewRBACMatcher", func(t *testing.T) {
		t.Parallel()

		path := writeRBACConfig(t, `
defaultPolicy: bogus
rules:
  - path: ""
`)
		cfg, err := middleware.LoadRBACConfigFromFile(path)
		require.NoError(t, err, "loader should only decode")

		_, err = middleware.NewRBACMatcher(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bogus")
	})
}
