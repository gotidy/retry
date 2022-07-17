# Retry operation with different strategies

Retry operations with constant, delays and exponential backoff strategies.

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev] [![Go Report Card](https://goreportcard.com/badge/github.com/gotidy/retry)][goreport]

[godev]: https://pkg.go.dev/github.com/gotidy/retry
[goreport]: https://goreportcard.com/report/github.com/gotidy/retry

## Installation

`go get github.com/gotidy/retry`

Requireg 1.18, because generics are used.

## Example

```go
    delays := Delays[1*time.Second, 2*time.Second, 4*time.Second]
	result, err := Retry(ctx, delays, func(ctx context.Context) (int, error) {
		return 0, errors.New("")
	})
```

## Documentation

[godev]

## License

[Apache 2.0](LICENSE)
