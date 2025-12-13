package logger

import (
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

// LogErrorWithContext логирует ошибку с контекстом
func LogErrorWithContext(err error, message string, fields ...zap.Field) {
	if err == nil {
		return
	}

	allFields := append(fields, zap.Error(err))
	Log.Error(message, allFields...)
}

// LogRetryAttempt логирует попытку повтора
func LogRetryAttempt(attempt int, delay time.Duration, err error) {
	Log.Warn("Retry attempt",
		zap.Int("attempt", attempt),
		zap.Duration("delay", delay),
		zap.Error(err),
	)
}

// LogRetrySuccess логирует успешное завершение после повторных попыток
func LogRetrySuccess(operation string, totalAttempts int) {
	Log.Info("Operation completed successfully after retries",
		zap.String("operation", operation),
		zap.Int("total_attempts", totalAttempts),
	)
}

// LogRetryFailure логирует окончательную неудачу после всех попыток
func LogRetryFailure(operation string, maxAttempts int, finalError error) {
	Log.Error("Operation failed after all retry attempts",
		zap.String("operation", operation),
		zap.Int("max_attempts", maxAttempts),
		zap.Error(finalError),
	)
}

// LogDatabaseError логирует ошибку базы данных с деталями
func LogDatabaseError(operation string, err error, query string, params ...interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Error(err),
	}

	if query != "" {
		fields = append(fields, zap.String("query", query))
	}

	if len(params) > 0 {
		fields = append(fields, zap.Any("parameters", params))
	}

	Log.Error("Database operation failed", fields...)
}

// LogNetworkError логирует сетевую ошибку
func LogNetworkError(operation string, url string, err error, statusCode int) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("url", url),
		zap.Error(err),
	}

	if statusCode > 0 {
		fields = append(fields, zap.Int("status_code", statusCode))
	}

	Log.Error("Network operation failed", fields...)
}

// LogBatchOperation логирует операции с батчами
func LogBatchOperation(operation string, batchSize int, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Int("batch_size", batchSize),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Log.Error("Batch operation failed", fields...)
	} else {
		Log.Info("Batch operation completed", fields...)
	}
}

// LogMetricUpdate логирует обновление метрик
func LogMetricUpdate(metricType string, name string, value interface{}, err error) {
	fields := []zap.Field{
		zap.String("metric_type", metricType),
		zap.String("metric_name", name),
		zap.Any("value", value),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Log.Error("Metric update failed", fields...)
	} else {
		Log.Debug("Metric updated successfully", fields...)
	}
}

// WithRequestID добавляет ID запроса к логам
func WithRequestID(requestID string) []zap.Field {
	return []zap.Field{zap.String("request_id", requestID)}
}

// WithComponent добавляет компонент к логам
func WithComponent(component string) []zap.Field {
	return []zap.Field{zap.String("component", component)}
}

// WithUserID добавляет ID пользователя к логам
func WithUserID(userID string) []zap.Field {
	return []zap.Field{zap.String("user_id", userID)}
}
