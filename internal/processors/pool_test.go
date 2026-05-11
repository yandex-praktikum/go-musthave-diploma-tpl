package processors_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/processors"
	"github.com/KalessinD/gophermart/internal/repositories/file/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// MockTaskProcessor создает мок-функцию обработки, которая записывает ID задач.
func mockTaskProcessor(t *testing.T, calls *[]string) processors.TaskProcessor {
	t.Helper()
	return func(ctx context.Context, _ chan time.Duration, task *processors.Task) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			*calls = append(*calls, task.ID)
			return nil
		}
	}
}

// errorTaskProcessor возвращает ошибку при обработке
func errorTaskProcessor() processors.TaskProcessor {
	return func(_ context.Context, _ chan time.Duration, _ *processors.Task) error {
		return errors.New("processing error")
	}
}

func TestWorkerPool_Processing(t *testing.T) {
	t.Run("process tasks in LIFO order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storageMock := mocks.NewMockPersistStorageInterface(ctrl)

		storageMock.EXPECT().Save(gomock.Any()).Return(nil).Times(1)

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		log := zap.NewNop()
		taskCh := make(chan *processors.Task, 10)
		pauseCh := make(chan time.Duration, 1)

		var processedTasks []string
		action := mockTaskProcessor(t, &processedTasks)

		pool, err := processors.NewQueueProcessor(1, 2, log, taskCh, pauseCh, storageMock, action)
		require.NoError(t, err)

		pool.Start(ctx)

		taskCh <- &processors.Task{ID: "1"}
		taskCh <- &processors.Task{ID: "2"}
		taskCh <- &processors.Task{ID: "3"}

		time.Sleep(100 * time.Millisecond)

		pool.Stop()
		pool.Wait()

		require.NotEmpty(t, processedTasks)
		require.Equal(t, 3, len(processedTasks))
	})

	t.Run("stop and wait gracefully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storageMock := mocks.NewMockPersistStorageInterface(ctrl)
		storageMock.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		log := zap.NewNop()
		taskCh := make(chan *processors.Task)
		pauseCh := make(chan time.Duration)

		var callCount int32
		action := func(_ context.Context, _ chan time.Duration, _ *processors.Task) error {
			atomic.AddInt32(&callCount, 1)
			return nil
		}

		pool, err := processors.NewQueueProcessor(2, 5, log, taskCh, pauseCh, storageMock, action)
		require.NoError(t, err)

		pool.Start(ctx)

		taskCh <- &processors.Task{ID: "stop-test"}

		time.Sleep(50 * time.Millisecond)
		require.Equal(t, int32(1), atomic.LoadInt32(&callCount))

		pool.Stop()
		pool.Wait()
	})

	t.Run("handle error in processor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storageMock := mocks.NewMockPersistStorageInterface(ctrl)
		storageMock.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()

		ctx := t.Context()
		log := zap.NewNop()

		taskCh := make(chan *processors.Task, 1)
		pauseCh := make(chan time.Duration, 1)

		pool, err := processors.NewQueueProcessor(1, 1, log, taskCh, pauseCh, storageMock, errorTaskProcessor())
		require.NoError(t, err)

		pool.Start(ctx)

		require.NotPanics(t, func() {
			taskCh <- &processors.Task{ID: "error-task"}
			time.Sleep(50 * time.Millisecond)
		})

		pool.Stop()
		pool.Wait()
	})

	t.Run("graceful shutdown dumps pending tasks", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storageMock := mocks.NewMockPersistStorageInterface(ctrl)

		ctx := t.Context()
		log := zap.NewNop()
		taskCh := make(chan *processors.Task, 10)
		pauseCh := make(chan time.Duration, 1)

		// Процессор не важен, так как воркеров не будет
		action := func(_ context.Context, _ chan time.Duration, _ *processors.Task) error {
			return nil
		}

		// poolSize = 0 (нет воркеров, которые могли бы "вытащить" задачу из канала)
		// bufSize = 0 (небуферизированный канал, отправка заблокируется)
		// Это заставит диспетчера "зависнуть" на передаче задачи, что позволит протестировать спасение задачи при Stop.
		pool, err := processors.NewQueueProcessor(0, 0, log, taskCh, pauseCh, storageMock, action)
		require.NoError(t, err)

		pool.Start(ctx)

		// Отправляем задачу. Она попадет в pendingTasks, диспетчер её вытащит,
		// попытается отправить в workerCh и заблокируется там навсегда (некому читать).
		taskCh <- &processors.Task{ID: "pending-task"}

		// Небольшая пауза, чтобы диспетчер успел дойти до блокировки отправки
		time.Sleep(50 * time.Millisecond)

		// Ожидаем, что Save будет вызван ровно 1 раз.
		// gomock.Len(1) проверяет, что в слайсе 1 элемент (наша необработанная задача)
		storageMock.EXPECT().Save(gomock.Len(1)).Return(nil).Times(1)

		pool.Stop()
		pool.Wait()
	})
}
