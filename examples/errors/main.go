package main

import (
	"fmt"

	"github.com/pkg/errors"

	keelerrors "github.com/foomo/keel/errors"
)

var (
	ErrOne = errors.New("one")
	ErrTwo = errors.New("two")
)

type ErrThree struct {
	error
}

func (e *ErrThree) Foo() string {
	return e.Error()
}

func main() {
	err1 := ErrOne
	err2 := keelerrors.NewWrappedError(err1, ErrTwo)
	err3 := &ErrThree{error: errors.New("error three")}
	err4 := keelerrors.NewWrappedError(err3, ErrTwo)
	err5 := keelerrors.NewWrappedError(ErrTwo, err3)

	if errors.Is(err1, ErrOne) {
		fmt.Println("err1 = ErrOne") //nolint:forbidigo
	}
	if errors.Is(err2, ErrTwo) {
		fmt.Println("err2 = ErrTwo") //nolint:forbidigo
	}
	if errors.Is(err2, ErrOne) {
		fmt.Println("err2 = ErrOne") //nolint:forbidigo
	}
	{
		var foo *ErrThree
		if errors.As(err3, &foo) {
			fmt.Println("err3 = ErrThree (" + foo.Foo() + ")") //nolint:forbidigo
		}
	}
	{
		var foo *ErrThree
		if errors.As(err4, &foo) {
			fmt.Println("err4 = ErrThree (" + foo.Foo() + ")") //nolint:forbidigo
		}
	}
	{
		var foo *ErrThree
		if errors.As(err5, &foo) {
			fmt.Println("err5 = ErrThree (" + foo.Foo() + ")") //nolint:forbidigo
		}
	}
}
