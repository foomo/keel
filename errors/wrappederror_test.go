package keelerrors_test

import (
	"fmt"
	"testing"

	keelerrors "github.com/foomo/keel/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleNewWrappedError() {
	parentErr := errors.New("parent")
	childErr := errors.New("child")
	wrappedErr := keelerrors.NewWrappedError(parentErr, childErr)
	fmt.Println(parentErr)
	fmt.Println(childErr)
	fmt.Println(wrappedErr)
	// Output:
	// parent
	// child
	// parent: child
}

func TestNewWrappedError(t *testing.T) {
	parentErr := errors.New("parent")
	childErr := errors.New("child")
	assert.Error(t, keelerrors.NewWrappedError(parentErr, childErr))
}

func TestWrapped(t *testing.T) {
	parentErr := errors.New("parent")
	childErr := errors.New("child")
	assert.Error(t, keelerrors.NewWrappedError(parentErr, childErr))
}

func Test_wrappedError_As(t *testing.T) {
	type (
		Parent struct {
			error
		}
		Child struct {
			error
		}
	)
	parentErr := &Parent{error: errors.New("parent")}
	childErr := &Child{error: errors.New("parent")}
	wrappedErr := keelerrors.NewWrappedError(parentErr, childErr)

	var (
		p *Parent
		c *Child
	)
	if assert.ErrorAs(t, wrappedErr, &p) {
		assert.EqualError(t, p, parentErr.Error())
	}
	if assert.ErrorAs(t, wrappedErr, &c) {
		assert.EqualError(t, c, childErr.Error())
	}
}

func Test_wrappedError_Error(t *testing.T) {
	parentErr := errors.New("parent")
	childErr := errors.New("child")
	wrappedErr := keelerrors.NewWrappedError(parentErr, childErr)
	assert.Equal(t, "parent: child", wrappedErr.Error())
}

func Test_wrappedError_Is(t *testing.T) {
	parentErr := errors.New("parent")
	childErr := errors.New("child")
	wrappedErr := keelerrors.NewWrappedError(parentErr, childErr)
	require.ErrorIs(t, wrappedErr, parentErr)
	require.ErrorIs(t, wrappedErr, childErr)
}
