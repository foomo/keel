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

func main() {

	err1 := ErrOne
	err2 := keelerrors.NewWrappedError(err1, ErrTwo)

	if errors.Is(err1, ErrOne) {
		fmt.Println("err1 = ErrOne")
	}
	if errors.Is(err2, ErrTwo) {
		fmt.Println("err2 = ErrTwo")
	}
	if errors.Is(err2, ErrOne) {
		fmt.Println("err2 = ErrOne")
	}
}
