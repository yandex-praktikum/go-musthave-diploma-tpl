package fcounter

import (
	"errors"
	"sync"
	"time"
)

type failure struct {
	timestamp time.Time
	err       error
}

type FailureCounter struct {
	mu          sync.RWMutex
	isFlushed   bool
	perDuration time.Duration
	ringPtr     int
	failures    []failure
}

// NewFailureCounter создает объект счетчика ошибок с пороговыми параметрами не более maxFailures ошибок в perDuration период
// Использование: добавляем ошибки в счетчик при каждом не успешном действии и сбрасываем счетчик при каждом успешном действии
func NewFailureCounter(maxFailures int, perDuration time.Duration) *FailureCounter {
	if maxFailures <= 0 {
		maxFailures = 3
	}
	if perDuration <= 100*time.Millisecond {
		perDuration = 100 * time.Millisecond
	}
	return &FailureCounter{
		isFlushed:   true,
		perDuration: perDuration,
		failures:    make([]failure, maxFailures),
	}
}

func (c *FailureCounter) isLimitExceeded() bool {
	return !c.isFlushed && time.Since(c.failures[c.ringPtr].timestamp) <= c.perDuration
}

// IsLimitExceeded возвращает количество актуальных ошибок
func (c *FailureCounter) IsLimitExceeded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.isLimitExceeded()
}

// Inc увеличивает счетчик ошибок и возвращает true, если порог превышен
func (c *FailureCounter) Inc(err error) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isFlushed = false
	c.failures[c.ringPtr].timestamp = time.Now()
	c.failures[c.ringPtr].err = err
	c.ringPtr = (c.ringPtr + 1) % len(c.failures)
	return c.isLimitExceeded()
}

// Reset сбрасывает счетчик ошибок в 0
func (c *FailureCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isFlushed {
		return
	}
	c.isFlushed = true
	size := len(c.failures)
	for range size {
		c.ringPtr = (c.ringPtr + size - 1) % size
		if time.Since(c.failures[c.ringPtr].timestamp) > c.perDuration {
			return
		}
		c.failures[c.ringPtr] = failure{}
	}
}

// Clear Реинициализирует объект полностью, создавая нагрузку на GC
func (c *FailureCounter) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isFlushed = true
	c.failures = make([]failure, len(c.failures))
	c.ringPtr = 0
}

// Возвращает актуальные ошибки
func (c *FailureCounter) Error() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isFlushed {
		return nil
	}
	size := len(c.failures)
	errs := make([]error, 0, size)
	for i := 1; i <= size; i++ {
		idx := (c.ringPtr + size - i) % size
		if time.Since(c.failures[idx].timestamp) > c.perDuration {
			return errors.Join(errs...)
		}
		errs = append(errs, c.failures[idx].err)
	}
	return errors.Join(errs...)
}
