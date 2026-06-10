package keeltemporal //nolint:testpackage

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"go.temporal.io/api/serviceerror"
)

func TestIsNamespaceNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "NamespaceNotFound",
			err:  serviceerror.NewNamespaceNotFound("order"),
			want: true,
		},
		{
			name: "NotFound",
			err:  serviceerror.NewNotFound("namespace missing"),
			want: true,
		},
		{
			name: "wrapped NamespaceNotFound",
			err:  errors.Wrap(serviceerror.NewNamespaceNotFound("order"), "x"),
			want: true,
		},
		{
			name: "Unavailable",
			err:  serviceerror.NewUnavailable("server down"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isNamespaceNotFound(tt.err); got != tt.want {
				t.Errorf("isNamespaceNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
