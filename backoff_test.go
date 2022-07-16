package retry

import (
	"reflect"
	"testing"
	"time"
)

func TestDelays(t *testing.T) {
	delays := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	next := Delays(delays...).Iterator()
	for _, want := range append(delays, StopDelay) {
		got := next()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: %v, want: %v", got, want)
		}
	}
}

func testRepeatly(t *testing.T, b BackOff, want time.Duration) {
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
	testRepeatly(t, Zero, 0)
}

func TestStop(t *testing.T) {
	testRepeatly(t, Stop, StopDelay)
}

func TestExponential(t *testing.T) {
	const muliplier = 1.5
	want := time.Second
	next := Exponential(want, muliplier).Iterator()
	for i := 0; i < 100; i++ {
		got := next()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: %v, want: %v", got, want)
		}
		want = time.Duration(float64(want) * muliplier)
	}
}
