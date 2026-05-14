package utils

import (
	"time"

	"github.com/avast/retry-go/v5"
	"go.uber.org/zap"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/logger"
)

func LinerDelay(n uint, err error, config retry.DelayContext) time.Duration {
	return config.Delay() + time.Duration(2*int(n))*time.Second
}

func NewRetryer(
	attempt uint, delay, maxDelay time.Duration, retryIf func(err error) bool) *retry.Retrier {
	options := []retry.Option{
		retry.Attempts(attempt),
		retry.Delay(delay),
		retry.MaxDelay(maxDelay),
		retry.LastErrorOnly(true),
		retry.DelayType(LinerDelay),
		retry.OnRetry(func(n uint, err error) {
			logger.Log.Warn("Retrying",
				zap.Uint("attempt", n),
				zap.Error(err),
			)
		}),
		retry.RetryIf(retryIf),
	}

	return retry.New(options...)
}
