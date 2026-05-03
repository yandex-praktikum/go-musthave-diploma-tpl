package processors_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/processors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewWorker(t *testing.T) {
	logger := zap.NewNop()
	ch := make(chan *processors.Task)
	pauseCh := make(chan time.Duration, 1)
	defer close(pauseCh)

	t.Run("success creation", func(t *testing.T) {
		worker, err := processors.NewWorker(ch, logger, func(_ context.Context, _ chan time.Duration, _ *processors.Task) error {
			return nil
		})

		require.NoError(t, err)
		require.NotNil(t, worker)
		require.NotEmpty(t, worker.ID())
	})
}

func TestWorker_Run(t *testing.T) {
	logger := zap.NewNop()

	t.Run("stops on context cancel", func(t *testing.T) {
		ch := make(chan *processors.Task)
		pauseCh := make(chan time.Duration, 1)
		defer close(pauseCh)

		// Мок процессора, который просто ждет вечно (если воркер не остановится, тест зависнет)
		action := func(_ context.Context, _ chan time.Duration, _ *processors.Task) error {
			return nil
		}

		worker, _ := processors.NewWorker(ch, logger, action)

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			worker.Run(ctx, pauseCh)
		}()

		// Даем воркеру время запуститься
		time.Sleep(10 * time.Millisecond)

		// Отменяем контекст
		cancel()

		// Ждем завершения с таймаутом
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Успех
		case <-time.After(1 * time.Second):
			t.Fatal("Worker did not stop after context cancel")
		}
	})

	t.Run("stops on channel close", func(t *testing.T) {
		ch := make(chan *processors.Task)
		pauseCh := make(chan time.Duration, 1)
		defer close(pauseCh)

		worker, _ := processors.NewWorker(ch, logger, nil)

		ctx := context.Background()
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			worker.Run(ctx, pauseCh)
		}()

		// Закрываем канал
		close(ch)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Успех
		case <-time.After(1 * time.Second):
			t.Fatal("Worker did not stop after channel close")
		}
	})

	t.Run("processes tasks correctly", func(t *testing.T) {
		ch := make(chan *processors.Task, 1)
		pauseCh := make(chan time.Duration, 1)
		defer close(pauseCh)

		var processedTask *processors.Task
		var mu sync.Mutex

		action := func(_ context.Context, _ chan time.Duration, task *processors.Task) error {
			mu.Lock()
			defer mu.Unlock()
			processedTask = task
			return nil
		}

		worker, _ := processors.NewWorker(ch, logger, action)

		task := &processors.Task{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go worker.Run(ctx, pauseCh)

		ch <- task

		require.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return processedTask != nil
		}, 1*time.Second, 10*time.Millisecond, "Task was not processed")
	})

	t.Run("handles processing error and continues", func(t *testing.T) {
		ch := make(chan *processors.Task, 2)
		pauseCh := make(chan time.Duration, 1)
		defer close(pauseCh)

		callCount := 0
		var mu sync.Mutex

		action := func(_ context.Context, _ chan time.Duration, _ *processors.Task) error {
			mu.Lock()
			defer mu.Unlock()
			callCount++
			if callCount == 1 {
				return errors.New("processing failed")
			}
			return nil
		}

		worker, _ := processors.NewWorker(ch, logger, action)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go worker.Run(ctx, pauseCh)

		ch <- &processors.Task{}
		ch <- &processors.Task{}

		require.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return callCount == 2
		}, 1*time.Second, 10*time.Millisecond, "Worker should continue processing after error")
	})
}
