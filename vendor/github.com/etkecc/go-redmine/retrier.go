package redmine

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

const (
	// MaxRetries is the maximum number of retries
	MaxRetries = 5
	// RetryDelay is the delay step between retries
	RetryDelay = 5 * time.Second
)

// Retry retries fn until it succeeds or MaxRetries is reached; use it for every call to the underlying API object.
func Retry(log *slog.Logger, fn func() (redmine.StatusCode, error)) error {
	_, err := retry(log, func() (struct{}, redmine.StatusCode, error) {
		statusCode, err := fn()
		return struct{}{}, statusCode, err
	})
	return err
}

// RetryResult retries fn until it succeeds or MaxRetries is reached; use for every underlying API object call.
func RetryResult[V any](log *slog.Logger, fn func() (V, redmine.StatusCode, error)) (V, error) {
	return retry(log, fn)
}

// retry is the shared retry loop behind Retry and RetryResult.
func retry[V any](log *slog.Logger, fn func() (V, redmine.StatusCode, error)) (V, error) {
	var zero V
	for i := 1; i <= MaxRetries; i++ {
		v, statusCode, err := fn()
		if statusCode == http.StatusNotFound {
			log.Warn("issue not found")
			return zero, nil
		}
		if statusCode > 499 {
			if i < MaxRetries {
				log.Warn("retrying", retryAttrs(i, statusCode, err)...)
				time.Sleep(RetryDelay * time.Duration(i))
				continue
			}
			log.Warn("giving up", retryAttrs(i, statusCode, err)...)
		}

		if err != nil {
			return v, err
		}

		return v, nil
	}
	return zero, fmt.Errorf("failed after %d retries", MaxRetries)
}

// retryAttrs builds the fields shared by the retry warnings, omitting a nil error.
func retryAttrs(attempt int, statusCode redmine.StatusCode, err error) []any {
	attrs := []any{slog.Int("retries", attempt), slog.Int("status_code", int(statusCode))}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}
	return attrs
}
