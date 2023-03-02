package retry

import (
	"fmt"
	"time"
)

// Wrapper of strategy.
type Wrapper interface {
	Wrap(s Strategy) Strategy
}

// MaxRetriesWrapper wraps the strategy with the max retrying stopper.
type MaxRetriesWrapper struct {
	MaxRetries int
	Strategy   Strategy
}

// MaxRetries wraps the strategy with the max retrying stopper.
func MaxRetries(n int, strategy Strategy) MaxRetriesWrapper {
	return MaxRetriesWrapper{MaxRetries: n, Strategy: strategy}
}

// MaxRetries wraps the strategy with the max retrying stopper.
// func WithMaxRetries(n int, strategy Strategy) MaxRetriesWrapper {
// 	return MaxRetriesWrapper{MaxRetries: n, Strategy: strategy}
// }

// Wrap other Strategy.
func (w MaxRetriesWrapper) Wrap(s Strategy) Strategy {
	return MaxRetriesWrapper{
		MaxRetries: w.MaxRetries,
		Strategy:   s,
	}
}

// Iterator returns an iterator that iterate over the inherited iterator and stops when the count of retries will be exhausted.
func (w MaxRetriesWrapper) Iterator() Iterator {
	n := 0
	iter := w.Strategy.Iterator()
	return func() (time.Duration, error) {
		if n >= w.MaxRetries {
			return StopDelay, fmt.Errorf("maximum retries elapsed: %d", w.MaxRetries)
		}
		n++
		return iter()
	}
}

// MaxElapsedTimeWrapper wraps the strategy with the max elapsed time stopper.
type MaxElapsedTimeWrapper struct {
	MaxElapsedTime time.Duration
	Strategy       Strategy
}

// MaxElapsedTime wraps the strategy with the max elapsed time stopper.
func MaxElapsedTime(d time.Duration, strategy Strategy) MaxElapsedTimeWrapper {
	return MaxElapsedTimeWrapper{MaxElapsedTime: d, Strategy: strategy}
}

// Wrap other Strategy.
func (w MaxElapsedTimeWrapper) Wrap(s Strategy) Strategy {
	return MaxElapsedTimeWrapper{
		MaxElapsedTime: w.MaxElapsedTime,
		Strategy:       s,
	}
}

// Iterator returns an iterator that iterate over the inherited iterator and stops when the time be elapsed.
func (w MaxElapsedTimeWrapper) Iterator() Iterator {
	start := time.Now()
	iter := w.Strategy.Iterator()
	return func() (time.Duration, error) {
		if time.Since(start) > w.MaxElapsedTime {
			return StopDelay, fmt.Errorf("retrying time elapsed: %s", w.MaxElapsedTime.String())
		}
		return iter()
	}
}
