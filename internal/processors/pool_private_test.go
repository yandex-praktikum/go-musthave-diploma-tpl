package processors

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/file"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testPoolSetup создает экземпляр WorkerPool для модульного тестирования приватных методов.
func testPoolSetup(t *testing.T, dumper file.PersistStorageInterface) *WorkerPool {
	t.Helper()

	return &WorkerPool{
		log:          zap.NewNop(),
		dumper:       dumper,
		pendingTasks: make(TaskList, 0),
		workerCh:     make(chan *Task, 10),
		hasTasks:     make(chan struct{}, 1), // Буфер 1 обязателен для RestoreQueue
	}
}

func TestWorkerPool_SleepForAWhile(t *testing.T) {
	t.Run("wait full duration", func(t *testing.T) {
		wp := &WorkerPool{}
		start := time.Now()
		delay := 50 * time.Millisecond

		wp.sleepForAWhile(context.Background(), delay)

		elapsed := time.Since(start)

		require.GreaterOrEqual(t, elapsed.Milliseconds(), delay.Milliseconds())
	})

	t.Run("interrupt by context", func(t *testing.T) {
		wp := &WorkerPool{}
		ctx, cancel := context.WithCancel(t.Context())

		// Отменяем контекст через 10мс
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		wp.sleepForAWhile(ctx, 1*time.Second)

		// Должны проснуться раньше
		require.Less(t, time.Since(start), 500*time.Millisecond)
	})
}

func TestWorkerPool_DumpQueue(t *testing.T) {
	t.Run("saves pending and buffered tasks", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "dump.json")
		storage, _ := file.NewJSONFileStorage(filePath)

		wp := testPoolSetup(t, storage)

		wp.pendingTasks = append(wp.pendingTasks, &Task{ID: "task_mem_1"})
		wp.pendingTasks = append(wp.pendingTasks, &Task{ID: "task_mem_2"})

		wp.workerCh <- &Task{ID: "task_chan_1"}

		wp.dumpQueue()

		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var orders models.OrdersList
		err = json.Unmarshal(data, &orders)
		require.NoError(t, err)

		require.Len(t, orders, 3)

		ids := make(map[string]bool)
		for _, o := range orders {
			ids[o.ID] = true
		}

		require.True(t, ids["task_mem_1"])
		require.True(t, ids["task_mem_2"])
		require.True(t, ids["task_chan_1"])
	})
}

func TestWorkerPool_RestoreQueue(t *testing.T) {
	t.Run("restore and erase", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "dump.json")
		storage, _ := file.NewJSONFileStorage(filePath)

		ordersToSave := models.OrdersList{
			{ID: "restore_1"},
			{ID: "restore_2"},
		}
		data, _ := json.Marshal(ordersToSave)
		err := os.WriteFile(filePath, data, 0o600)

		require.NoError(t, err)

		wp := testPoolSetup(t, storage)
		require.Empty(t, wp.pendingTasks)

		wp.RestoreQueue()

		require.Len(t, wp.pendingTasks, 2)
		require.Equal(t, "restore_1", wp.pendingTasks[0].ID)
		require.Equal(t, "restore_2", wp.pendingTasks[1].ID)

		select {
		case <-wp.hasTasks:
			// Успех, сигнал пришел
		default:
			t.Fatal("Ожидался сигнал в канал hasTasks после восстановления")
		}

		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "Файл дампа должен быть удален")
	})

	t.Run("no file no panic", func(t *testing.T) {
		dir := t.TempDir()
		storage, _ := file.NewJSONFileStorage(filepath.Join(dir, "missing.json"))
		wp := testPoolSetup(t, storage)

		wp.RestoreQueue()
		require.Empty(t, wp.pendingTasks)

		select {
		case <-wp.hasTasks:
			t.Fatal("Сигнал не должен был прийти при пустом восстановлении")
		default:
		}
	})
}
