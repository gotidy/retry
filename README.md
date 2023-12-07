# Retry operation with different strategies

Retry operations with constant, delays and exponential backoff strategies.

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev] [![Go Report Card](https://goreportcard.com/badge/github.com/gotidy/retry)][goreport]

[godev]: https://pkg.go.dev/github.com/gotidy/retry
[goreport]: https://goreportcard.com/report/github.com/gotidy/retry

## Installation

`go get github.com/gotidy/retry`

Required at least 1.18 version of Go compiler.

## Example

### Custom delays

```go
delays := retry.Delays[1*time.Second, 2*time.Second, 4*time.Second]
result, err := retry.DoR(ctx, delays, func(ctx context.Context) (int, error) {
    return 0, errors.New("")
})
```

### Exponential

```go
err := retry.Do(ctx, retry.Exponential(time.Second, 1.5, 0.5), func(ctx context.Context) (int, error) {
    return 0, errors.New("")
})
```

```go
result, err := retry.DoR(ctx, retry.TruncatedExponential(time.Second, 1.5, 0, 10*time.Second), func(ctx context.Context) (int, error) {
    return 0, errors.New("")
})
```

There are also other strategies such as Constant, Zero.

### Permanent error

If need to prevent retrying wrap error with `Permanent``.

```go
result, err := retry.DoR(ctx, retry.Constant(time.Second), func(ctx context.Context) (int, error) {
    result, err := DoSomething()
    switch {
    case errors.Is(err, ErrInvalidArguments)  
        return retry.Permanent(err) // Prevent retrying.  
    default:
        return err
    }
    return result, nil
})
```

[More](/examples_test.go)

## Documentation

[godev]

## License

[Apache 2.0](LICENSE)
