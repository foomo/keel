package testing_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/foomo/keel/messaging/testing"
)

func ExampleGoSync() {
	ctx := context.Background()
	fmt.Println("1. starting")

	testing.GoSync(ctx, func(ctx context.Context) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("2. go routine started")
	})

	fmt.Println("3. complete")

	// Output:
	// 1. starting
	// 2. go routine started
	// 3. complete
}

func ExampleGoSyncE() {
	ctx := context.Background()
	fmt.Println("1. starting")

	var ErrCustom = errors.New("custom error")

	err := testing.GoSyncE(ctx, func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("2. go routine started")
		return ErrCustom
	})

	fmt.Println("3. complete:", errors.Is(err, context.Canceled), errors.Is(err, ErrCustom))

	// Output:
	// 1. starting
	// 2. go routine started
	// 3. complete: false true
}
