package retry

import (
	"reflect"
	"testing"
	"time"
)

func TestDelays(t *testing.T) {
	delays := Delays{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	next := delays.Iterator()
	for _, want := range append(delays, StopDelay) {
		got := next()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: %v, want: %v", got, want)
		}
	}
}

func testRepeatly(t *testing.T, b Strategy, want time.Duration) {
	next := b.Iterator()
	for i := 0; i < 10; i++ {
		got := next()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: %v, want: %v", got, want)
		}
	}
}

func TestConstant(t *testing.T) {
	delay := time.Second
	testRepeatly(t, Constant(delay), delay)
}

func TestZero(t *testing.T) {
	testRepeatly(t, Zero(), 0)
}

func TestStop(t *testing.T) {
	testRepeatly(t, Stop(), StopDelay)
}

func TestExponential(t *testing.T) {
	testExponentialBackOff(t, Exponential(time.Second, 1.5, 0))
}

func TestExponentialBackOff(t *testing.T) {
	testExponentialBackOff(t, Exponential(time.Second, 1.5, 0.5))
}

func TestTruncatedExponentialBackOff(t *testing.T) {
	testExponentialBackOff(t, TruncatedExponential(time.Second, 1.5, 0, 10*time.Second))
}

func testExponentialBackOff(t *testing.T, exp ExponentialBackOff) {
	delay := exp.Start
	next := exp.Iterator()
	for i := 0; i < 10; i++ {
		got := next()

		if exp.Jitter != 0 {
			minDelay := delay - time.Duration(exp.Jitter*float64(delay))
			maxDelay := delay + time.Duration(exp.Jitter*float64(delay))
			if !(minDelay <= got && got <= maxDelay) {
				t.Errorf("expected between %v and %v, got %v", minDelay, maxDelay, got)
			}
		} else if !reflect.DeepEqual(got, delay) {
			t.Errorf("got: %v, want: %v", got, delay)
		}
		delay = time.Duration(float64(delay) * exp.Factor)
		if exp.MaxDelay != 0 && delay > exp.MaxDelay {
			delay = exp.MaxDelay
		}
	}
}

func TestJitter(t *testing.T) {
	// 33% chance of being 1.
	assertEquals(t, 1, jitter(2, 0.5, 0))
	assertEquals(t, 1, jitter(2, 0.5, 0.33))
	// 33% chance of being 2.
	assertEquals(t, 2, jitter(2, 0.5, 0.34))
	assertEquals(t, 2, jitter(2, 0.5, 0.66))
	// 33% chance of being 3.
	assertEquals(t, 3, jitter(2, 0.5, 0.67))
	assertEquals(t, 3, jitter(2, 0.5, 0.99))
}

func assertEquals[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
}
