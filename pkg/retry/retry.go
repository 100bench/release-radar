package retry

import (
	"time"

	"github.com/mackb/releaseradar/pkg/logger"
)

// Do retries a function with exponential backoff.
func Do(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			logger.L().Sugar().Warnf("retrying after error: %v", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = fn()
		if err == nil {
			return nil
		}
	}
	return err
}
