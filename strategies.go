package retry

import (
	"errors"
	"math/rand"
	"time"
)

var ErrDelaysSpent = errors.New("all delays spent")
var ErrStopped = errors.New("stopped")

// StopDelay indicates that no more retries should be made.
const StopDelay time.Duration = -1

// Iterator is a delays generator.
type Iterator func() (time.Duration, error)

// Strategy is a retrying strategy for retrying an operation.
type Strategy interface {
	Iterator() Iterator
}

// Delays is a retry strategy that returns specified delays.
type Delays []time.Duration

// Iterator returns the specified delays generator.
func (d Delays) Iterator() Iterator {
	i := 0
	return func() (time.Duration, error) {
		if i >= len(d) {
			return StopDelay, ErrDelaysSpent
		}
		current := d[i]
		i++
		return current, nil
	}
}

// Constant is a retry strategy that always returns the same retry delay.
type Constant time.Duration

// Iterator returns constant delay generator.
func (c Constant) Iterator() Iterator {
	return func() (time.Duration, error) {
		if time.Duration(c) == StopDelay {
			return time.Duration(c), ErrStopped
		}
		return time.Duration(c), nil
	}
}

// Zero is zero delayed strategy is a fixed retry strategy whose retry time is always zero,
// meaning that the operation is retried immediately without waiting, indefinitely.
func Zero() Constant {
	return Constant(0)
}

// Stop is a fixed retry strategy that always returns StopDelay,
// meaning that the operation should never be retried.
func Stop() Constant {
	return Constant(StopDelay)
}

// ExponentialBackOff is exponential backoff strategy.
type ExponentialBackOff struct {
	// Start delay.
	Start time.Duration
	// Multiplier factor. Next delay = delay * multiplier.
	Factor float64
	// Delay randomization. delay = delay * (random value in range [1 - Jitter, 1 + Jitter]).
	Jitter float64
	// Delay maximum.
	MaxDelay time.Duration
}

// Exponential creates exponential backoff strategy.
func Exponential(start time.Duration, factor float64, jitter float64) ExponentialBackOff {
	return ExponentialBackOff{Start: start, Factor: factor, Jitter: jitter}
}

// TruncatedExponential creates exponential backoff strategy with max delay.
func TruncatedExponential(start time.Duration, factor, jitter float64, maxDelay time.Duration) ExponentialBackOff {
	return ExponentialBackOff{Start: start, Factor: factor, Jitter: jitter, MaxDelay: maxDelay}
}

// Iterator returns exponential backoff delays generator.
func (e ExponentialBackOff) Iterator() Iterator {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	delay := e.Start
	return func() (time.Duration, error) {
		cur := delay
		delay = time.Duration(float64(delay) * e.Factor)
		if e.MaxDelay != 0 && delay >= e.MaxDelay {
			delay = e.MaxDelay
		}
		return jitter(cur, e.Jitter, rand.Float64()), nil
	}
}

func jitter(delay time.Duration, factor, random float64) time.Duration {
	if factor == 0 {
		return delay
	}
	delta := factor * float64(delay)
	minDelay := float64(delay) - delta
	maxDelay := float64(delay) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minDelay + (random * (maxDelay - minDelay + 1)))
}
