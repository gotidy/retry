package retry

import (
	"testing"
	"time"
)

func TestMaxRetries(t *testing.T) {
	const maxRetries = 5
	iter := MaxRetries(maxRetries, Zero()).Iterator()
	n := 0
	for iter() != StopDelay {
		n++
	}
	if n != maxRetries {
		t.Errorf("tries want: %d, got: %d", maxRetries, n)
	}
}

func TestMaxElapsedTime(t *testing.T) {
	const maxElapsedTime = 3 * time.Second
	start := time.Now()
	iter := MaxElapsedTime(maxElapsedTime, Zero()).Iterator()
	for iter() != StopDelay {
		time.Sleep(time.Second)
	}
	if d := time.Since(start); d < maxElapsedTime {
		t.Errorf("expected %s >= %s", d, maxElapsedTime)
	}
}
