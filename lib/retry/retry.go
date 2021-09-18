package retry

import (
	"fmt"
	"time"
)

// Function signature of retryable function
type RetryableFunc func() error

// Retry call f every interval until the maximum number of attempts is reached.
// If the incoming attempts is 0, retry forever
func Retry(attempts int, interval time.Duration, f RetryableFunc) (err error) {
	for i := 0; ; i++ {
		err = f()

		if err == nil {
			return
		}

		if attempts != 0 && i >= (attempts-1) {
			break
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("after %d attempts, last error: %v", attempts, err)
}
