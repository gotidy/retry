package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDoWithResult(t *testing.T) {
	const retries = 5
	const want = 10
	count := 0

	got, err := DoR(context.Background(), Zero(), func(ctx context.Context) (int, error) {
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

func TestDoWithResult_Error(t *testing.T) {
	delays := Delays{time.Second, time.Second}
	count := 0

	_, err := DoR(context.Background(), delays, func(ctx context.Context) (int, error) {
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
	_, err := DoR(ctx, Zero(), func(ctx context.Context) (int, error) {
		cancel()
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}

	_, err = DoR(ctx, Constant(time.Second), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	})
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetryWithResult_Timeout(t *testing.T) {
	_, err := DoR(context.Background(), Constant(100*time.Millisecond), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	}, WithTimeout(time.Second))
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetryWithResult_MaxRetrying(t *testing.T) {
	const retries = 5
	count := 0
	_, err := DoRN(context.Background(), Zero(), func(ctx context.Context) (int, error) {
		count++
		return 0, errors.New("")
	}, retries)
	if err == nil {
		t.Error("expected error but got nil")
	}
	if count != retries {
		t.Errorf("unexpected count of retries: %d, expected: %d", count, retries)
	}
}

func TestRetryWithResult_MaxElapsedTime(t *testing.T) {
	_, err := DoRE(context.Background(), Constant(100*time.Millisecond), func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	}, time.Second)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestRetry(t *testing.T) {
	const retries = 5
	count := 0

	err := Do(context.Background(), Zero(), func(ctx context.Context) error {
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

func TestPermanent(t *testing.T) {
	want := errors.New("want")
	err := Permanent(want)

	got := errors.Unwrap(err)
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if fmt.Sprintf("%s", want) != fmt.Sprintf("%s", err) {
		t.Errorf("Error() = %v, want %v", err.Error(), want.Error())
	}
}

func TestDo_PermanentError(t *testing.T) {
	const retries = 5
	count := 0
	wantErr := errors.New("error")

	err := Do(context.Background(), Zero(), func(ctx context.Context) error {
		count++
		return Permanent(wantErr)
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error: %s, got: %s", wantErr, err)
	}
	if count != 1 {
		t.Errorf("unexpected count of retries: %d, expected: %d", count, 1)
	}
}

func TestDo_Notify(t *testing.T) {
	const retries = 2
	const constDelay = time.Microsecond

	_ = Do(context.Background(), Constant(constDelay), func(ctx context.Context) error {
		return errors.New("error")
	}, WithNotify(func(err error, delay time.Duration, try int, elapsed time.Duration) {
		if delay != constDelay {
			t.Errorf("want delay: %s, got: %s", constDelay, delay)
		}
		if try != 1 {
			t.Errorf("want try: %d, got: %d", 1, try)
		}
	}), WithMaxRetries(retries))
}

func TestAs(t *testing.T) {
	e := &Error{}
	if got := As(e); e != got {
		t.Errorf("As() = %v, want %v", got, e)
	}

	if got := As(errors.New("error")); got != nil {
		t.Errorf("As() = %v, want nil", got)
	}
}

func TestUnwrap(t *testing.T) {
	err := errors.New("error")
	if got := Unwrap(&Error{Err: err}); got != err {
		t.Errorf("As() = %v, want %v", got, err)
	}
}
