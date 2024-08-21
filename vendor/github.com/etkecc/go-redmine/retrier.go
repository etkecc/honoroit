package redmine

import (
	"fmt"
	"net/http"
	"time"

	redmine "github.com/nixys/nxs-go-redmine/v5"
	"github.com/rs/zerolog"
)

const (
	// MaxRetries is the maximum number of retries
	MaxRetries = 5
	// RetryDelay is the delay step between retries
	RetryDelay = 5 * time.Second
)

// Retry retries the given function until it succeeds or the maximum number of retries is reached
// it should be used everywhere you call the underlying API object (inside the library: r.cfg.api, outside: r.GetAPI())
func Retry(log *zerolog.Logger, fn func() (redmine.StatusCode, error)) error {
	for i := 1; i <= MaxRetries; i++ {
		statusCode, err := fn()
		if statusCode == http.StatusNotFound {
			log.Warn().Msg("issue not found")
			return nil
		}
		if statusCode > 499 {
			if i < MaxRetries {
				log.Warn().Int("retries", i).Int("status_code", int(statusCode)).Err(err).Msg("retrying")
				time.Sleep(RetryDelay * time.Duration(i))
				continue
			}
			log.Warn().Int("retries", i).Int("status_code", int(statusCode)).Err(err).Msg("giving up")
		}

		if err != nil {
			return err
		}

		return nil
	}
	return fmt.Errorf("failed after %d retries", MaxRetries)
}

// RetryResult retries the given function until it succeeds or the maximum number of retries is reached
// it should be used everywhere you call the underlying API object (inside the library: r.cfg.api, outside: r.GetAPI())
func RetryResult[V any](log *zerolog.Logger, fn func() (V, redmine.StatusCode, error)) (V, error) {
	var zero V
	for i := 1; i <= MaxRetries; i++ {
		v, statusCode, err := fn()
		if statusCode == http.StatusNotFound {
			log.Warn().Msg("issue not found")
			return zero, nil
		}
		if statusCode > 499 {
			if i < MaxRetries {
				log.Warn().Int("retries", i).Int("status_code", int(statusCode)).Err(err).Msg("retrying")
				time.Sleep(RetryDelay * time.Duration(i))
				continue
			}
			log.Warn().Int("retries", i).Int("status_code", int(statusCode)).Err(err).Msg("giving up")
		}

		if err != nil {
			return v, err
		}

		return v, nil
	}
	return zero, fmt.Errorf("failed after %d retries", MaxRetries)
}
