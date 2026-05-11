package clients_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/KalessinD/gophermart/internal/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrServiceIsBusy2_Error(t *testing.T) {
	delay := 5 * time.Second
	cause := errors.New("internal connection error")

	t.Run("error message without cause", func(t *testing.T) {
		err := clients.NewErrServiceIsBusy(delay, nil)
		expectedMsg := "service responds about too many requests, retry after 5s"
		assert.EqualError(t, err, expectedMsg)
	})

	t.Run("error message with cause", func(t *testing.T) {
		err := clients.NewErrServiceIsBusy(delay, cause)
		expectedMsg := "service responds about too many requests, retry after 5s: internal connection error"
		assert.EqualError(t, err, expectedMsg)
	})
}

func TestErrServiceIsBusy2_GetDelay(t *testing.T) {
	delay := 10 * time.Second
	err := clients.NewErrServiceIsBusy(delay, nil)

	assert.Equal(t, delay, err.GetDelay())
}

func TestErrServiceIsBusy2_errors_Is(t *testing.T) {
	err := clients.NewErrServiceIsBusy(1*time.Second, nil)

	t.Run("matches sentinel error", func(t *testing.T) {
		assert.True(t, errors.Is(err, clients.ErrServiceIsBusy), "error should match ErrServiceIsBusy")
	})

	t.Run("does not match other errors", func(t *testing.T) {
		assert.False(t, errors.Is(err, errors.New("other")), "error should not match random error")
	})

	t.Run("matches inside wrapper", func(t *testing.T) {
		wrapped := fmt.Errorf("request failed: %w", err)
		assert.True(t, errors.Is(wrapped, clients.ErrServiceIsBusy), "errors.Is should find sentinel inside wrapped error")
	})
}

func TestErrServiceIsBusy2_errors_As(t *testing.T) {
	delay := 15 * time.Second
	cause := errors.New("timeout")
	err := clients.NewErrServiceIsBusy(delay, cause)

	t.Run("extracts custom error", func(t *testing.T) {
		var target *clients.ServiceIsBusyError
		ok := errors.As(err, &target)
		require.True(t, ok, "errors.As should find ErrServiceIsBusy2")
		assert.Equal(t, delay, target.GetDelay())
	})

	t.Run("extracts from wrapper", func(t *testing.T) {
		wrapped := fmt.Errorf("context: %w", err)
		var target *clients.ServiceIsBusyError
		ok := errors.As(wrapped, &target)
		require.True(t, ok, "errors.As should find ErrServiceIsBusy2 inside wrapper")
		assert.Equal(t, delay, target.GetDelay())
	})
}

func TestErrServiceIsBusy2_Unwrap(t *testing.T) {
	cause := errors.New("underlying issue")
	err := clients.NewErrServiceIsBusy(1*time.Second, cause)

	t.Run("returns cause via Unwrap", func(t *testing.T) {
		unwrapped := errors.Unwrap(err)
		assert.EqualError(t, unwrapped, "underlying issue")
	})

	t.Run("errors.Is finds cause via chain", func(t *testing.T) {
		assert.True(t, errors.Is(err, cause), "errors.Is should find the cause error")
	})
}

func TestErrServiceIsBusy2_errors_Join(t *testing.T) {
	err1 := clients.NewErrServiceIsBusy(5*time.Second, nil)
	err2 := errors.New("another error")

	joined := errors.Join(err1, err2)

	t.Run("identifies sentinel in joined error", func(t *testing.T) {
		assert.True(t, errors.Is(joined, clients.ErrServiceIsBusy))
	})

	t.Run("identifies other error in joined error", func(t *testing.T) {
		assert.True(t, errors.Is(joined, err2))
	})

	t.Run("extracts custom error from joined", func(t *testing.T) {
		var target *clients.ServiceIsBusyError
		ok := errors.As(joined, &target)
		require.True(t, ok)
		assert.Equal(t, 5*time.Second, target.GetDelay())
	})
}
