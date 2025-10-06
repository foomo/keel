package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// NewResource creates and returns a default resource for telemetry data, using the provided context.
func NewResource(ctx context.Context) (*resource.Resource, error) {
	var attrs []attribute.KeyValue

	envs := map[attribute.Key][]string{
		semconv.VCSRepositoryNameKey:    {"REPO_NAME", "GIT_REPOSITORY_NAME", "GITHUB_REPOSITORY_NAME", "GIT_OTEL_VCS_REPOSITORY_NAME"},
		semconv.VCSRepositoryURLFullKey: {"REPO_URL", "GIT_REPOSITORY_URL", "GITHUB_REPOSITORY", "OTEL_VCS_REPOSITORY_URL_FULL"},
		semconv.VCSRefBaseNameKey:       {"OTEL_VCS_BASE_NAME"},
		semconv.VCSRefBaseRevisionKey:   {"OTEL_VCS_BASE_REVSION"},
		semconv.VCSRefBaseTypeKey:       {"OTEL_VCS_BASE_TYPE"},
		semconv.VCSRefHeadNameKey:       {"GIT_BRANCH", "OTEL_VCS_HEAD_NAME"},
		semconv.VCSRefHeadRevisionKey:   {"GIT_COMMIT_HASH", "OTEL_VCS_HEAD_REVSION"},
		semconv.VCSRefHeadTypeKey:       {"OTEL_VCS_HEAD_TYPE"},
	}
	for k, keys := range envs {
		for _, key := range keys {
			if v := os.Getenv(key); v != "" {
				attrs = append(attrs, k.String(v))
				break
			}
		}
	}

	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(attrs...),
	)
}
