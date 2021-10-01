package backoff

import "time"

// Retry the function f until it does not return error or BackOff stops.
//
// Example:
// 	operation := func() error {
// 		// An operation that may fail
// 	}
//
// 	err := backoff.Retry(operation, backoff.NewExponentialBackoff())
// 	if err != nil {
// 		// handle error
// 	}
//
// 	// operation is successfull
func Retry(f func() error, b BackOff) error { return RetryNotify(f, b, nil) }

// RetryNotify calls notify function with the error and wait duration for each failed attempt before sleep.
func RetryNotify(f func() error, b BackOff, notify func(err error, wait time.Duration)) error {
	var err error
	var next time.Duration

	b.Reset()
	for {
		if err = f(); err == nil {
			return nil
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}

		time.Sleep(next)
	}
}
