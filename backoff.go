package retry

import "time"

const StopDelay time.Duration = -1

type Iterator func() time.Duration

type BackOff interface {
	Iterator() Iterator
}

type DelaysBackOff []time.Duration

func Delays(delays ...time.Duration) DelaysBackOff {
	return DelaysBackOff(delays)
}

func (d DelaysBackOff) Iterator() Iterator {
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

type ConstantBackOff time.Duration

func Constant(delay time.Duration) ConstantBackOff {
	return ConstantBackOff(delay)
}

func (c ConstantBackOff) Iterator() Iterator {
	return func() time.Duration {
		return time.Duration(c)
	}
}

var Zero = ZeroBackOff{}

type ZeroBackOff struct{}

func (z ZeroBackOff) Iterator() Iterator {
	return func() time.Duration {
		return 0
	}
}

var Stop = StopBackOff{}

func (z StopBackOff) Iterator() Iterator {
	return func() time.Duration {
		return StopDelay
	}
}

type StopBackOff struct{}

type ExponentialBackOff struct {
	Start     time.Duration
	Muliplier float64
}

func Exponential(start time.Duration, muliplier float64) ExponentialBackOff {
	return ExponentialBackOff{Start: start, Muliplier: muliplier}
}

func (e ExponentialBackOff) Iterator() Iterator {
	delay := e.Start
	return func() time.Duration {
		cur := delay
		delay = time.Duration(float64(delay) * e.Muliplier)
		return cur
	}
}
