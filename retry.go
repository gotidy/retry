package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/lib/ptr"
)

type options struct {
	// MaxRetries is maximum count of retries.
	MaxRetries int
	// After Timeout the context will canceled.
	Timeout time.Duration
	// After RetryingTimeElapse the retrying will stopped.
	RetryingTimeElapse time.Duration
}

// Option is a retring option setter.
type Option func(opts *options)

// WithMaxRetries sets maximum retries.
func WithMaxRetries(n int) Option {
	return func(opts *options) {
		opts.MaxRetries = n
	}
}

// WithTimeout sets timemeout.
func WithTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.Timeout = d
	}
}

// WithRetryingTimeElapse sets RetryingTimeElapse option.
// Time after which retrying are stopped.
func WithRetryingTimeElapse(d time.Duration) Option {
	return func(opts *options) {
		opts.RetryingTimeElapse = d
	}
}

// RetryWithResult function with specified back off.
func RetryWithResult[T any](ctx context.Context, strategy Strategy, operation func(ctx context.Context) (T, error), o ...Option) (result T, err error) {
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

	next := strategy.Iterator()
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

// Retry function with specified back off.
func Retry(ctx context.Context, strategy Strategy, operation func(ctx context.Context) error, o ...Option) (err error) {
	_, err = RetryWithResult(ctx, strategy, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, operation(ctx)
	}, o...)
	return err
}
