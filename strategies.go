package retry

import (
	"math/rand"
	"time"
)

// StopDelay indicates that no more retries should be made.
const StopDelay time.Duration = -1

// Iterator is a delays generator.
type Iterator func() time.Duration

// Strategy is a backoff policy for retrying an operation.
type Strategy interface {
	Iterator() Iterator
}

// Delays is a back off policy that returns specified delays.
type Delays []time.Duration

// Iterator returns the specified delays iterator.
func (d Delays) Iterator() Iterator {
	i := 0
	return func() time.Duration {
		if i >= len(d) {
			return StopDelay
		}
		current := d[i]
		i++
		return current
	}
}

// Constant returns constant retrying delay.
type Constant time.Duration

// Iterator returns constant delay generator.
func (c Constant) Iterator() Iterator {
	return func() time.Duration {
		return time.Duration(c)
	}
}

// Zero is zero delayed strategy.
func Zero() Constant {
	return Constant(0)
}

// Stop is stop strategy.
func Stop() Constant {
	return Constant(StopDelay)
}

// ExponentialBackOffStrategy returns exponential increased delays.
type ExponentialBackOffStrategy struct {
	// start delay
	Start time.Duration
	// Multiplier factor. Next delay = delay * multiplier
	Factor float64
	// Delay randomization. delay = delay * (random value in range [1 - Jitter, 1 + Jitter])
	Jitter   float64
	MaxDelay time.Duration
}

// Exponential creates exponential strategy.
func Exponential(start time.Duration, factor float64) ExponentialBackOffStrategy {
	return ExponentialBackOffStrategy{Start: start, Factor: factor}
}

// ExponentialBackOff creates exponential back off.
func ExponentialBackOff(start time.Duration, factor float64, jitter float64) ExponentialBackOffStrategy {
	return ExponentialBackOffStrategy{Start: start, Factor: factor, Jitter: jitter}
}

// TruncatedExponentialBackOff creates exponential back off.
func TruncatedExponentialBackOff(start time.Duration, factor, jitter float64, maxDelay time.Duration) ExponentialBackOffStrategy {
	return ExponentialBackOffStrategy{Start: start, Factor: factor, Jitter: jitter, MaxDelay: maxDelay}
}

// Iterator returns exponential delays generator.
func (e ExponentialBackOffStrategy) Iterator() Iterator {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	delay := e.Start
	return func() time.Duration {
		cur := delay
		delay = time.Duration(float64(delay) * e.Factor)
		if e.MaxDelay != 0 && delay >= e.MaxDelay {
			delay = e.MaxDelay
		}
		return jitter(cur, e.Jitter, rand.Float64())
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
