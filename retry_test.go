package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithResult(t *testing.T) {
	const retries = 5
	const want = 10
	count := 0

	got, err := RetryWithResult(context.Background(), Zero(), func(ctx context.Context) (int, error) {
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

func TestRetryWithResult_Error(t *testing.T) {
	delays := Delays{time.Second, time.Second}
	count := 0

	_, err := RetryWithResult(context.Background(), delays, func(ctx context.Context) (int, error) {
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

func TestRetryWithResult_Context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := RetryWithResult(ctx, Zero(), func(ctx context.Context) (int, error) {
		cancel()
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}

	_, err = RetryWithResult(ctx, Constant(time.Second), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetryWithResult_Timeout(t *testing.T) {
	_, err := RetryWithResult(context.Background(), Constant(100*time.Millisecond), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	}, WithTimeout(time.Second))
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetryWithResult_MaxRetring(t *testing.T) {
	const retries = 5
	count := 0
	_, err := RetryWithResult(context.Background(), Zero(), func(ctx context.Context) (int, error) {
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

func TestRetryWithResult_RetryingTimeElapse(t *testing.T) {
	_, err := RetryWithResult(context.Background(), Constant(100*time.Millisecond), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	}, WithRetryingTimeElapse(time.Second))
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetry(t *testing.T) {
	const retries = 5
	count := 0

	err := Retry(context.Background(), Zero(), func(ctx context.Context) error {
		if count == retries {
			return nil
		}
		count++
		return errors.New("")
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != retries {
		t.Errorf("unexpected count of retries: %d, expected: %d", count, retries)
	}
}
