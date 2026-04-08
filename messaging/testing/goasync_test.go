package testing_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/foomo/keel/messaging/testing"
)

func ExampleGoAsync() {
	ctx := context.Background()
	fmt.Println("1. starting")

	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("2. go routine started")
		done()
	})

	fmt.Println("2. waiting for go routine")
	<-await

	fmt.Println("4. complete")

	// Output:
	// 1. starting
	// 2. waiting for go routine
	// 2. go routine started
	// 4. complete
}

func ExampleGoAsyncE() {
	ctx := context.Background()
	fmt.Println("1. starting")

	var ErrCustom = errors.New("custom error")

	errCh := testing.GoAsyncE(ctx, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("2. go routine started")
		return ErrCustom
	})

	fmt.Println("2. waiting for go routine")
	err := <-errCh

	fmt.Println("4. complete:", errors.Is(err, ErrCustom))

	// Output:
	// 1. starting
	// 2. waiting for go routine
	// 2. go routine started
	// 4. complete: true
}
