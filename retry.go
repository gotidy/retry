package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/lib/ptr"
)

type options struct {
	MaxRetries         int
	Timeout            time.Duration
	RetryingTimeElapse time.Duration
}

type Option func(opts *options)

func WithMaxRetries(n int) Option {
	return func(opts *options) {
		opts.MaxRetries = n
	}
}

func WithTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.Timeout = d
	}
}

func WithRetryingTimeElapse(d time.Duration) Option {
	return func(opts *options) {
		opts.RetryingTimeElapse = d
	}
}

func Retry[T any](ctx context.Context, backOff BackOff, operation func(ctx context.Context) (T, error), o ...Option) (result T, err error) {
	var opts options
	for _, opt := range o {
		opt(&opts)
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

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
			return nil
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

		if opts.RetryingTimeElapse > 0 && opts.RetryingTimeElapse <= time.Since(start) {
			return ptr.Zero[T](), fmt.Errorf("retring time elapsed: %s: %w", opts.RetryingTimeElapse, err)
		}

		if opts.MaxRetries > 0 && opts.MaxRetries <= retrying {
			return ptr.Zero[T](), fmt.Errorf("maximum retries elapsed: %d: %w", opts.MaxRetries, err)
		}

		delay := next()
		if delay == StopDelay {
			return ptr.Zero[T](), fmtErr()
		}

		select {
		case <-ctx.Done():
			return ptr.Zero[T](), fmtErr()
		case <-time.After(delay):
		}

		retrying++
	}
}
