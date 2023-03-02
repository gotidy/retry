package retry

import (
	"testing"
	"time"
)

func TestMaxRetries(t *testing.T) {
	t.Parallel()

	const maxRetries = 5
	iter := MaxRetries(maxRetries, Zero()).Iterator()
	n := 0
	for {
		if d, err := iter(); d == StopDelay || err != nil {
			break
		}
		n++
	}
	if n != maxRetries {
		t.Errorf("tries want: %d, got: %d", maxRetries, n)
	}
}

func TestMaxElapsedTime(t *testing.T) {
	t.Parallel()

	const maxElapsedTime = 3 * time.Second
	start := time.Now()
	iter := MaxElapsedTime(maxElapsedTime, Zero()).Iterator()
	for {
		if d, err := iter(); d == StopDelay || err != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if d := time.Since(start); d < maxElapsedTime {
		t.Errorf("expected %s >= %s", d, maxElapsedTime)
	}
}
