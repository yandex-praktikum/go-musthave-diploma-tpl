package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewWorker(t *testing.T) {
	svc, _ := newOrderFixture(t, nil)
	w := NewWorker(svc)

	if w.orderService != svc {
		t.Error("worker should hold the passed order service")
	}
	if w.batchSize != 100 {
		t.Errorf("batchSize = %d, want 100", w.batchSize)
	}
	if w.interval != 5*time.Second {
		t.Errorf("interval = %v, want 5s", w.interval)
	}
	if w.ticker == nil {
		t.Error("ticker should be initialized")
	}

	w.ticker.Stop()
}

func TestWorker_Start_StopsOnContextCancel(t *testing.T) {
	svc, mock := newOrderFixture(t, NewAccrualClient("http://127.0.0.1:1"))

	// Первая обработка при старте воркера: запрос ожидающих заказов.
	mock.ExpectQuery(`SELECT number, user_id, status, accrual, uploaded_at, updated_at FROM orders`).
		WillReturnRows(sqlmockPendingRows())

	w := NewWorker(svc)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	// Даём воркеру выполнить первую итерацию, затем останавливаем.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// воркер корректно завершился
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop after context cancel")
	}
}

func TestWorker_Stop(t *testing.T) {
	svc, _ := newOrderFixture(t, nil)
	w := NewWorker(svc)

	// Stop закрывает канал остановки и останавливает тикер.
	w.Stop()
	w.ticker.Stop()
}

// sqlmockPendingRows возвращает пустой набор строк ожидающих заказов.
func sqlmockPendingRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at", "updated_at"})
}
