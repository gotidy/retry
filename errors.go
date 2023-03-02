package retry

import (
	"errors"
	"fmt"
	"time"
)

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
