package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func ExampleDoR() {
	count := 0

	n, err := DoR(context.Background(), Constant(time.Millisecond), func(ctx context.Context) (int, error) {
		if count == 5 {
			return count, nil
		}
		count++
		return 0, errors.New("not yet")
	})
	fmt.Println(n)
	fmt.Println(err)
	// Output:
	// 5
	// <nil>
}

func ExampleDo() {
	err := Do(context.Background(), Constant(time.Millisecond), func(ctx context.Context) error {
		return Permanent(errors.New("oops"))
	})
	fmt.Println(err)
	// Output:
	// oops
}

func ExampleExponential() {
	count := 1
	err := Do(context.Background(), Exponential(time.Second, 2, 0), func(ctx context.Context) error {
		if count > 3 {
			return nil
		}
		count++
		return errors.New("don't touch me")
	}, WithNotify(func(err error, delay time.Duration, try int, elapsed time.Duration) {
		fmt.Println(delay, try)
	}))
	fmt.Println(err)
	// Output:
	// 1s 1
	// 2s 2
	// 4s 3
	// <nil>
}
