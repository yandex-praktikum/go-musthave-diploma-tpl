package retry

import (
	"github.com/avast/retry-go/v4"
	"time"
)

func MakeRetry(retryableFunc retry.RetryableFunc) error {
	return retry.Do(
		retryableFunc,
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			switch n {
			case 0:
				return 1 * time.Second
			case 1:
				return 3 * time.Second
			case 2:
				return 5 * time.Second
			default:
				return retry.BackOffDelay(n, err, config)
			}
		}),
		retry.Attempts(4),
	)
}
