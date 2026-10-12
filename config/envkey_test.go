package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/foomo/keel/config"
)

func TestAutomaticEnvKeyMapping(t *testing.T) {
	for _, tt := range []struct{ name, key, env, wrongEnv string }{
		{"dots", "keel.test.category.translation.url", "KEEL_TEST_CATEGORY_TRANSLATION_URL", "KEEL_TEST_CATEGORYTRANSLATION_URL"},
		{"camel case", "keel.test.categoryTranslation.url", "KEEL_TEST_CATEGORYTRANSLATION_URL", "KEEL_TEST_CATEGORY_TRANSLATION_URL"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.env, "")
			t.Setenv(tt.wrongEnv, "wrong")
			assert.Empty(t, config.Config().GetString(tt.key), "camel case does not introduce underscores")
			t.Setenv(tt.env, "configured")
			assert.Equal(t, "configured", config.Config().GetString(tt.key))
		})
	}
}
