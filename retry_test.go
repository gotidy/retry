package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	const retries = 5
	const want = 10
	count := 0

	got, err := Retry(context.Background(), Zero, func(ctx context.Context) (int, error) {
		if count == retries {
			return want, nil
		}
		count++
		return 0, errors.New("")
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != retries {
		t.Errorf("unexpected count of retries: %d, expected: %d", count, retries)
	}
	if got != want {
		t.Errorf("value got: %v, want: %v", got, want)
	}
}

func TestRetry_Error(t *testing.T) {
	delays := []time.Duration{time.Second, time.Second}
	count := 0

	_, err := Retry(context.Background(), Delays(delays...), func(ctx context.Context) (int, error) {
		count++
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}
	if count != len(delays)+1 {
		t.Errorf("unexpected count of retries: %s", err)
	}
}

func TestRetry_Context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := Retry(ctx, Zero, func(ctx context.Context) (int, error) {
		cancel()
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}

	_, err = Retry(ctx, Constant(time.Second), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetry_Timeout(t *testing.T) {
	_, err := Retry(context.Background(), Constant(100*time.Millisecond), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	}, WithTimeout(time.Second))
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetry_MaxRetring(t *testing.T) {
	const retries = 5
	count := 0
	_, err := Retry(context.Background(), Zero, func(ctx context.Context) (int, error) {
		count++
		return 0, errors.New("")
	}, WithMaxRetries(retries))
	if err == nil {
		t.Error("expected error but got nil")
	}
	if count != retries {
		t.Errorf("unexpected count of retries: %d, expected: %d", count, retries)
	}
}

func TestRetry_RetryingTimeElapse(t *testing.T) {
	_, err := Retry(context.Background(), Constant(100*time.Millisecond), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	}, WithRetryingTimeElapse(time.Second))
	if err == nil {
		t.Error("expected error but got nil")
	}
}
