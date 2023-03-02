package retry

import (
	"context"
	"errors"
	"fmt"
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
		delay, nErr := next()
		elapsed := time.Since(start)
		if delay == StopDelay {
			return ptr.Zero[T](), newError(err, ctx.Err(), nErr.Error(), retrying, delay, elapsed)
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

// Do retries the operation with specified strategy.
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

// PermanentError signals that the operation should not be retried.
type PermanentError struct {
	Err error
}

// Permanent wrap error with permanent error.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return PermanentError{Err: err}
}

func (e PermanentError) Error() string {
	return e.Err.Error()
}

func (e PermanentError) Unwrap() error {
	return e.Err
}

// Error wraps the original error and contains information about the last retry.
type Error struct {
	LastDelay   time.Duration
	ElapsedTime time.Duration
	Retries     int
	Msg         string
	Err         error
}

func newError(err, ctxErr error, msg string, retries int, lastDelay time.Duration, elapsed time.Duration) error {
	e := &Error{
		ElapsedTime: elapsed,
		Retries:     retries,
		LastDelay:   lastDelay,
		Err:         err,
	}
	switch {
	case ctxErr != nil && err == nil:
		e.Msg = fmt.Sprintf("retrying %d canceled, time elapsed: %s, last delay: %s", retries, elapsed, lastDelay)
		e.Err = ctxErr
	case ctxErr != nil && err != nil:
		e.Msg = fmt.Sprintf("retrying %d canceled: %s, time elapsed: %s, last delay: %s", retries, ctxErr.Error(), elapsed, lastDelay)
	case ctxErr == nil && err != nil:
		e.Msg = fmt.Sprintf("retrying %d stopped, time elapsed: %s, last delay: %s", retries, elapsed, lastDelay)
	default:
		return nil
	}
	if msg != "" {
		e.Msg = msg + ": " + e.Msg
	}
	return e
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

// As returns retry Error that wrap an original operation error.
func As(err error) *Error {
	e := &Error{}
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// Unwrap returns an original operation error.
func Unwrap(err error) error {
	if e := As(err); e != nil {
		return e.Err
	}
	return err
}
