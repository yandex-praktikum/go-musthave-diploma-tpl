package fcounter

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFailureCounter(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		fc := NewFailureCounter(0, 0)
		assert.Equal(t, 3, len(fc.failures))
		assert.Equal(t, 100*time.Millisecond, fc.perDuration)
		assert.True(t, fc.isFlushed)
	})

	t.Run("custom values", func(t *testing.T) {
		fc := NewFailureCounter(5, time.Second)
		assert.Equal(t, 5, len(fc.failures))
		assert.Equal(t, time.Second, fc.perDuration)
	})
}

func TestFailureCounter_Inc_And_Limit(t *testing.T) {
	fc := NewFailureCounter(2, 500*time.Millisecond)

	// 1-я ошибка
	exceeded := fc.Inc(errors.New("err 1"))
	assert.False(t, exceeded)
	assert.False(t, fc.isFlushed)

	// 2-я ошибка (лимит достигнут: 2 ошибки за 500мс)
	exceeded = fc.Inc(errors.New("err 2"))
	assert.True(t, exceeded)
	assert.True(t, fc.IsLimitExceeded())
}

func TestFailureCounter_SlidingWindow(t *testing.T) {
	fc := NewFailureCounter(2, 100*time.Millisecond)

	fc.Inc(errors.New("err 1"))
	time.Sleep(150 * time.Millisecond) // ошибка "протухла"

	fc.Inc(errors.New("err 2"))
	// Хоть ошибок и 2 в буфере, первая уже не валидна
	assert.False(t, fc.IsLimitExceeded(), "Limit should not be exceeded because first error expired")
}

func TestFailureCounter_Reset(t *testing.T) {
	fc := NewFailureCounter(3, time.Hour)

	fc.Inc(errors.New("1"))
	fc.Inc(errors.New("2"))
	assert.False(t, fc.isFlushed)

	fc.Reset()
	assert.True(t, fc.isFlushed)
	assert.False(t, fc.IsLimitExceeded())
	assert.Nil(t, fc.Error())

	// Проверка return при повторном Reset (покрытие ветки if c.isFlushed)
	fc.Reset()
	assert.True(t, fc.isFlushed)
}

func TestFailureCounter_PartialReset(t *testing.T) {
	duration := 200 * time.Millisecond
	fc := NewFailureCounter(3, duration)

	fc.Inc(errors.New("old"))
	time.Sleep(duration + 10*time.Millisecond)
	fc.Inc(errors.New("new 1"))
	fc.Inc(errors.New("new 2"))

	// Reset должен дойти до "old", увидеть, что она старая, и сделать return
	fc.Reset()

	err := fc.Error()
	assert.Nil(t, err, "All fresh errors should be cleared")
}

func TestFailureCounter_Clear(t *testing.T) {
	fc := NewFailureCounter(3, time.Hour)
	fc.Inc(errors.New("err"))

	fc.Clear()

	assert.True(t, fc.isFlushed)
	assert.Equal(t, 0, fc.ringPtr)
	assert.Nil(t, fc.Error())
	// Проверка, что слайс пересоздан (все timestamp zero)
	for _, f := range fc.failures {
		assert.True(t, f.timestamp.IsZero())
	}
}

func TestFailureCounter_ErrorAggregation(t *testing.T) {
	fc := NewFailureCounter(3, time.Hour)
	e1 := errors.New("error 1")
	e2 := errors.New("error 2")

	fc.Inc(e1)
	fc.Inc(e2)

	err := fc.Error()
	require.Error(t, err)
	// errors.Join соединяет через \n. Проверяем наличие обеих строк
	assert.Contains(t, err.Error(), "error 1")
	assert.Contains(t, err.Error(), "error 2")
}

func TestFailureCounter_Concurrency(t *testing.T) {
	fc := NewFailureCounter(100, time.Hour)
	var wg sync.WaitGroup
	workers := 10

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				fc.Inc(errors.New("fail"))
				fc.IsLimitExceeded()
				fc.Error()
			}
		}()
	}

	wg.Wait()
	// Если нет panic или data race (запускать с -race), тест пройден
}
