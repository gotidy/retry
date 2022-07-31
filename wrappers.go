package retry

import (
	"time"
)

// MaxRetriesWrapper wraps the strategy with the max retrying stopper.
type MaxRetriesWrapper struct {
	MaxRetries int
	Strategy   Strategy
}

// MaxRetries wraps the strategy with the max retrying stopper.
func MaxRetries(n int, strategy Strategy) MaxRetriesWrapper {
	return MaxRetriesWrapper{MaxRetries: n, Strategy: strategy}
}

// Iterator returns an iterator that iterate over the inherited iterator and stops when the count of retries will be exhausted.
func (w MaxRetriesWrapper) Iterator() Iterator {
	n := 0
	iter := w.Strategy.Iterator()
	return func() time.Duration {
		if n == w.MaxRetries {
			return StopDelay
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

// Iterator returns an iterator that iterate over the inherited iterator and stops when the time be elapsed.
func (w MaxElapsedTimeWrapper) Iterator() Iterator {
	start := time.Now()
	iter := w.Strategy.Iterator()
	return func() time.Duration {
		if time.Since(start) > w.MaxElapsedTime {
			return StopDelay
		}
		return iter()
	}
}
