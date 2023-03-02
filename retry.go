// Operations retries with different strategies.
package retry

import (
	"context"
	"errors"
	"time"

	"github.com/gotidy/lib/ptr"
)

// Notify is a notify-on-error function. It receives an operation error and
// strategy delay if the operation failed (with an error).
//
// NOTE that if the strategy stated to stop retrying,
// the notify function isn't called.
type Notify func(err error, delay time.Duration, try int, elapsed time.Duration)

type options struct {
	// MaxRetries is maximum count of retries.
	MaxRetries int
	// After Timeout the context will canceled.
	Timeout time.Duration
	// After MaxElapsedTime the retrying will stopped.
	MaxElapsedTime time.Duration
	// Notify
	Notify Notify

	Strategy Strategy
}

// Option is a retrying option setter.
type Option func(opts *options)

// WithMaxRetries sets maximum retries.
func WithMaxRetries(n int) Option {
	return func(opts *options) {
		opts.Strategy = MaxRetriesWrapper{MaxRetries: n}.Wrap(opts.Strategy)
	}
}

// WithTimeout sets timeout.
func WithTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.Timeout = d
	}
}

// WithMaxElapsedTime sets MaxElapsedTime option.
// Time after which retrying are stopped.
func WithMaxElapsedTime(d time.Duration) Option {
	return func(opts *options) {
		opts.Strategy = MaxElapsedTimeWrapper{MaxElapsedTime: d}.Wrap(opts.Strategy)
	}
}

// WithNotify sets maximum retries.
func WithNotify(n Notify) Option {
	return func(opts *options) {
		opts.Notify = n
	}
}

// DoR retries the operation with result and specified strategy.
// To stop the retry, the operation must return a permanent error, see Permanent(err).
func DoR[T any](ctx context.Context, strategy Strategy, operation func(ctx context.Context) (T, error), o ...Option) (result T, err error) {
	opts := options{Strategy: strategy}

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
	var delay time.Duration

	next := opts.Strategy.Iterator()
	for {
		if ctx.Err() != nil {
			return ptr.Zero[T](), newError(err, ctx.Err(), "", retrying, delay, time.Since(start))
		}

		result, err = operation(ctx)
		if err == nil {
			return result, nil
		}
		var perm PermanentError
		if ok := errors.As(err, &perm); ok {
			return ptr.Zero[T](), perm
		}

		var nErr error
		prevDelay := delay
		delay, nErr = next()
		elapsed := time.Since(start)
		if delay == StopDelay {
			return ptr.Zero[T](), newError(err, ctx.Err(), nErr.Error(), retrying, prevDelay, elapsed)
		}

		if opts.Notify != nil {
			opts.Notify(err, delay, retrying, elapsed)
		}

		select {
		case <-ctx.Done():
			return ptr.Zero[T](), newError(err, ctx.Err(), "", retrying, delay, elapsed)
		case <-time.After(delay):
		}

		retrying++
	}
}

// DoRN retries the operation with result with specified strategy and the maximum number of retries.
func DoRN[T any](ctx context.Context, strategy Strategy, operation func(ctx context.Context) (T, error), maxReties int, o ...Option) (result T, err error) {
	return DoR(ctx, strategy, operation, append(o, WithMaxRetries(maxReties))...)
}

// DoRE retries the operation with result with specified strategy and maximum elapsed time.
func DoRE[T any](ctx context.Context, strategy Strategy, operation func(ctx context.Context) (T, error), maxElapsedTime time.Duration, o ...Option) (result T, err error) {
	return DoR(ctx, strategy, operation, append(o, WithMaxElapsedTime(maxElapsedTime))...)
}

// DoRNE retries the operation with result with the specified strategy and the maximum number of retries and maximum elapsed time.
func DoRNE[T any](ctx context.Context, strategy Strategy, operation func(ctx context.Context) (T, error), maxReties int, maxElapsedTime time.Duration, o ...Option) (result T, err error) {
	return DoR(ctx, strategy, operation, append(o, WithMaxRetries(maxReties), WithMaxElapsedTime(maxElapsedTime))...)
}

// Do retries the operation with specified strategy.
// To stop the retry, the operation must return a permanent error, see Permanent(err).
func Do(ctx context.Context, strategy Strategy, operation func(ctx context.Context) error, o ...Option) (err error) {
	_, err = DoR(ctx, strategy, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, operation(ctx)
	}, o...)
	return err
}

// DoN retries the operation with specified strategy and the maximum number of retries.
func DoN(ctx context.Context, strategy Strategy, operation func(ctx context.Context) error, maxReties int, o ...Option) (err error) {
	return Do(ctx, strategy, operation, append(o, WithMaxRetries(maxReties))...)
}

// DoE retries the operation with the specified strategy and maximum elapsed time.
func DoE(ctx context.Context, strategy Strategy, operation func(ctx context.Context) error, maxElapsedTime time.Duration, o ...Option) (err error) {
	return Do(ctx, strategy, operation, append(o, WithMaxElapsedTime(maxElapsedTime))...)
}

// DoNE retries the operation with the specified strategy and the maximum number of retries and maximum elapsed time.
func DoNE(ctx context.Context, strategy Strategy, operation func(ctx context.Context) error, maxReties int, maxElapsedTime time.Duration, o ...Option) (err error) {
	return Do(ctx, strategy, operation, append(o, WithMaxRetries(maxReties), WithMaxElapsedTime(maxElapsedTime))...)
}
