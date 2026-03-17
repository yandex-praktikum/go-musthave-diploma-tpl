package pgc

import (
	"context"
	"testing"
	"time"

	"gmart/internal/adapters/metrics"
	"gmart/internal/adapters/pgc/backoff"
	"gmart/internal/adapters/pgc/fcounter"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPgInstance_OnlineOffline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockPgInstanceMetrics(ctrl)
	mockPool := NewMockpgxPoolDriverIface(ctrl)

	h := &pgInstance{
		instanceName: "test_db",
		pgPool:       mockPool,
		metrics:      mockMetrics,
	}

	t.Run("Transition to Online", func(t *testing.T) {
		mockMetrics.EXPECT().SetStatus("test_db", true).Times(1)
		h.Online()
		assert.True(t, h.IsReady())
	})

	t.Run("Transition to Offline", func(t *testing.T) {
		mockMetrics.EXPECT().SetStatus("test_db", false).Times(1)
		mockMetrics.EXPECT().IncOfflineEvent("test_db").Times(1)
		h.Offline()
		assert.False(t, h.IsReady())
	})
}

func TestPgInstance_Tx_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockPgInstanceMetrics(ctrl)
	mockPool := NewMockpgxPoolDriverIface(ctrl)
	mockTx := NewMockTx(ctrl)

	h := &pgInstance{
		instanceName: "test_db",
		pgPool:       mockPool,
		metrics:      mockMetrics,
		failures:     fcounter.NewFailureCounter(3, time.Second),
		repeater:     backoff.NewPgBackoff(1, time.Millisecond),
	}
	h.isReady.Store(true)

	ctx := context.Background()

	mockPool.EXPECT().Begin(ctx).Return(mockTx, nil)
	mockTx.EXPECT().Commit(gomock.Any()).Return(nil)
	mockMetrics.EXPECT().ObserveLatency("test_db", metrics.OpTx, gomock.Any())
	mockMetrics.EXPECT().SetStatus("test_db", true).AnyTimes() // Из-за HandleError(nil)

	err := h.Tx(ctx, func(ctx context.Context, tx PgxTxIface) error {
		return nil
	})

	assert.NoError(t, err)
}

func TestPgInstance_HandleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := NewMockPgInstanceMetrics(ctrl)
	h := &pgInstance{
		instanceName: "test_db",
		metrics:      mockMetrics,
		failures:     fcounter.NewFailureCounter(3, time.Second),
	}
	h.isReady.Store(true)

	t.Run("Nil error keeps it online", func(t *testing.T) {
		mockMetrics.EXPECT().SetStatus("test_db", true).AnyTimes()
		err := h.HandleError(nil) // Теперь Reset() не упадет
		assert.NoError(t, err)
	})
}

func TestPgInstance_CanTry(t *testing.T) {
	h := &pgInstance{}
	h.isReady.Store(false)
	h.lastRetry.Store(0)

	t.Run("Allow first try in offline", func(t *testing.T) {
		assert.True(t, h.CanTry())
	})

	t.Run("Block subsequent tries immediately", func(t *testing.T) {
		h.lastRetry.Store(time.Now().Unix())
		assert.False(t, h.CanTry())
	})
}
