package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/lib/ptr"
)

func Retry[T any](ctx context.Context, backOff BackOff, operation func(ctx context.Context) (T, error)) (result T, err error) {
	start := time.Now()
	retrying := 1

	fmtErr := func() error {
		elapsed := time.Since(start)
		switch {
		case ctx.Err() != nil && err == nil:
			return fmt.Errorf("retrying \"%d\" canceled, time elapsed: %s: %w", retrying, elapsed, ctx.Err())
		case ctx.Err() != nil && err != nil:
			return fmt.Errorf("retrying \"%d\" canceled: %s, time elapsed: %s: %w", retrying, ctx.Err().Error(), elapsed, err)
		case ctx.Err() == nil && err != nil:
			return fmt.Errorf("retrying \"%d\" stoped, time elapsed: %s: %w", retrying, elapsed, err)
		default:
			panic("no errors")
		}
	}

	next := backOff.Iterator()
	for {
		if ctx.Err() != nil {
			return ptr.Zero[T](), fmtErr()
		}

		result, err = operation(ctx)
		if err == nil {
			return result, nil
		}

		delay := next()
		if delay == StopDelay {
			break
		}

		select {
		case <-ctx.Done():
			return ptr.Zero[T](), fmtErr()
		case <-time.After(delay):
		}
		retrying++
	}
	return ptr.Zero[T](), fmtErr()
}
