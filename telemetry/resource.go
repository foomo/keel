package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

// envAttributes maps environment variables to semantic attributes.
var envAttributes = map[attribute.Key][]string{
	semconv.ServiceNamespaceKey:     {"OTEL_SERVICE_NAMESPACE"},
	semconv.VCSRepositoryNameKey:    {"REPO_NAME", "REPOSITORY_NAME", "GIT_REPOSITORY_NAME", "GITHUB_REPOSITORY", "GITHUB_REPOSITORY_NAME", "GIT_OTEL_VCS_REPOSITORY_NAME"},
	semconv.VCSRepositoryURLFullKey: {"REPO_URL", "REPOSITORY_URL", "GIT_REPOSITORY_URL", "OTEL_VCS_REPOSITORY_URL_FULL"},
	semconv.VCSRefBaseNameKey:       {"OTEL_VCS_BASE_NAME"},
	semconv.VCSRefBaseRevisionKey:   {"OTEL_VCS_BASE_REVISION"},
	semconv.VCSRefBaseTypeKey:       {"OTEL_VCS_BASE_TYPE"},
	semconv.VCSRefHeadNameKey:       {"GIT_BRANCH", "OTEL_VCS_HEAD_NAME"},
	semconv.VCSRefHeadRevisionKey:   {"GIT_COMMIT", "GIT_COMMIT_HASH", "OTEL_VCS_HEAD_REVISION"},
	semconv.VCSRefHeadTypeKey:       {"GIT_TYPE", "OTEL_VCS_HEAD_TYPE"},
	"vcs.repository.path":           {"REPO_PATH", "REPOSITORY_PATH", "GIT_REPOSITORY_PATH", "OTEL_VCS_ROOT_PATH"},
}

// EnvAttributes returns the service namespace and VCS attributes read from
// well known environment variables. For each attribute the first non-empty
// variable wins.
func EnvAttributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	for k, keys := range envAttributes {
		for _, key := range keys {
			if v := os.Getenv(key); v != "" {
				attrs = append(attrs, k.String(v))
				break
			}
		}
	}

	return attrs
}

// NewResource returns the resource attached to all providers of this
// package. It combines OTEL_SERVICE_NAME, OTEL_RESOURCE_ATTRIBUTES and
// [EnvAttributes].
func NewResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(EnvAttributes()...),
	)
}
