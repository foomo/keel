package rbac_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/keel/net/http/rbac"
	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	testingx.Tags(t, tagx.Short)
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	reflector := new(jsonschema.Reflector)
	reflector.FieldNameTag = "yaml"
	reflector.RequiredFromJSONSchemaTags = true
	require.NoError(t, reflector.AddGoComments("github.com/foomo/keel/net/http/rbac", "./"))
	schema := reflector.Reflect(&rbac.Config{})
	schema.Title = "RBAC middleware configuration"
	schema.Description = "Path-based access-control rules consumed by middleware.RBAC. Authoritative validation lives in rbac.NewMatcher; this schema covers structural validation for editors."
	schema.ID = "https://github.com/foomo/keel/net/http/rbac/rbac.schema.json"
	actual, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	filename := path.Join(cwd, "rbac.schema.json")

	expected, err := os.ReadFile(filename)
	if !errors.Is(err, os.ErrNotExist) {
		require.NoError(t, err)
	}

	if !assert.Equal(t, string(expected), string(actual)) {
		require.NoError(t, os.WriteFile(filename, actual, 0600))
	}
}
