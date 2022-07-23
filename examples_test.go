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
