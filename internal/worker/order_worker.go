package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"go.uber.org/zap"
)

// ==================== Константы и переменные ====================

const (
	defaultWorkers         = 3
	defaultInterval        = 3 * time.Second
	defaultDBCheckInterval = 5 * time.Second
	defaultOrderChanSize   = 100
	defaultResultChanSize  = 100
	pollTimeout            = 10 * time.Second
)

// ==================== Структуры ====================

// PollResult представляет результат опроса заказа
type PollResult struct {
	OrderNumber string
	OrderInfo   *orderclient.OrderResponse
	Error       error
}

// OrderWorker интерфейс для воркера заказов
type OrderWorker interface {
	Start(ctx context.Context) error
	Stop()
	AddOrder(ctx context.Context, orderNumber string, interval time.Duration) error
	RemoveOrder(orderNumber string) error
	Results() <-chan PollResult
	GetActiveOrders() []string
	GetOrderStats() map[entity.OrderStatus]int
	ReloadOrdersFromDB(ctx context.Context) error
}

// orderTask структура задачи для отслеживания заказа
type orderTask struct {
	orderNumber string
	cancel      context.CancelFunc
	lastPolled  time.Time
	status      entity.OrderStatus
}

// orderWorker реализация OrderWorker
type orderWorker struct {
	client          *orderclient.Client
	repo            *repository.Repository
	logger          *zap.Logger
	workers         int
	interval        time.Duration
	dbCheckInterval time.Duration

	tasks      map[string]*orderTask
	mu         sync.RWMutex
	wg         sync.WaitGroup
	stopChan   chan struct{}
	isRunning  bool
	resultChan chan PollResult
	orderChan  chan string
}

// Config конфигурация воркера
type Config struct {
	Client          *orderclient.Client
	Repo            *repository.Repository
	Logger          *zap.Logger
	Workers         int
	Interval        time.Duration
	DBCheckInterval time.Duration
	OrderChanSize   int
	ResultChanSize  int
}

// ==================== Фабричные функции ====================

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig(client *orderclient.Client, repo *repository.Repository, logger *zap.Logger) Config {
	return Config{
		Client:          client,
		Repo:            repo,
		Logger:          logger,
		Workers:         defaultWorkers,
		Interval:        defaultInterval,
		DBCheckInterval: defaultDBCheckInterval,
		OrderChanSize:   defaultOrderChanSize,
		ResultChanSize:  defaultResultChanSize,
	}
}

// NewOrderWorker создает нового воркера для заказов
func NewOrderWorker(config Config) OrderWorker {
	// Устанавливаем значения по умолчанию
	if config.Workers <= 0 {
		config.Workers = defaultWorkers
	}
	if config.Interval <= 0 {
		config.Interval = defaultInterval
	}
	if config.DBCheckInterval <= 0 {
		config.DBCheckInterval = defaultDBCheckInterval
	}
	if config.OrderChanSize <= 0 {
		config.OrderChanSize = defaultOrderChanSize
	}
	if config.ResultChanSize <= 0 {
		config.ResultChanSize = defaultResultChanSize
	}

	// Если логгер не передан, создаем no-op логгер
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	return &orderWorker{
		client:          config.Client,
		repo:            config.Repo,
		logger:          config.Logger,
		workers:         config.Workers,
		interval:        config.Interval,
		dbCheckInterval: config.DBCheckInterval,
		tasks:           make(map[string]*orderTask),
		stopChan:        make(chan struct{}),
		resultChan:      make(chan PollResult, config.ResultChanSize),
		orderChan:       make(chan string, config.OrderChanSize),
	}
}

// ==================== Управление жизненным циклом ====================

// Start запускает воркер
func (w *orderWorker) Start(ctx context.Context) error {
	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	w.isRunning = true
	w.logger.Info("Starting order worker",
		zap.Int("workers", w.workers),
		zap.Duration("interval", w.interval),
		zap.Duration("db_check_interval", w.dbCheckInterval),
	)

	// Загружаем активные заказы из БД при старте
	if err := w.ReloadOrdersFromDB(ctx); err != nil {
		w.isRunning = false
		w.logger.Error("Failed to load orders from DB", zap.Error(err))
		return fmt.Errorf("failed to load orders from DB: %w", err)
	}

	// Запускаем обработчики
	w.startOrderHandlers(ctx)
	w.startDBScheduler(ctx)
	w.startOrderDistributor(ctx)

	w.logger.Info("Order worker started successfully",
		zap.Int("initial_tasks", len(w.tasks)),
	)
	return nil
}

// Stop останавливает воркер
func (w *orderWorker) Stop() {
	if !w.isRunning {
		return
	}

	w.logger.Info("Stopping order worker...",
		zap.Int("active_tasks", len(w.tasks)),
	)

	// Сигнализируем остановку
	close(w.stopChan)

	// Отменяем все задачи
	w.cancelAllTasks()

	// Ждем завершения всех горутин
	w.wg.Wait()

	// Закрываем каналы
	close(w.resultChan)

	w.isRunning = false
	w.logger.Info("Order worker stopped successfully")
}

// cancelAllTasks отменяет все активные задачи
func (w *orderWorker) cancelAllTasks() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for orderNumber, task := range w.tasks {
		task.cancel()
		delete(w.tasks, orderNumber)
		w.logger.Debug("Canceled task",
			zap.String("order_number", orderNumber),
			zap.String("status", string(task.status)),
		)
	}
}

// ==================== Управление заказами ====================

// AddOrder добавляет заказ для отслеживания
func (w *orderWorker) AddOrder(ctx context.Context, orderNumber string, interval time.Duration) error {
	if orderNumber == "" {
		w.logger.Error("Order number is empty")
		return fmt.Errorf("order number cannot be empty")
	}

	w.logger.Debug("Adding order to worker",
		zap.String("order_number", orderNumber),
		zap.Duration("interval", interval),
	)

	// Проверяем существование заказа в БД
	order, err := w.repo.Order().GetByNumber(ctx, orderNumber)
	if err != nil {
		w.logger.Error("Failed to check order in DB",
			zap.String("order_number", orderNumber),
			zap.Error(err),
		)
		return fmt.Errorf("failed to check order in DB: %w", err)
	}

	if order == nil {
		w.logger.Warn("Order not found in database",
			zap.String("order_number", orderNumber),
		)
		return fmt.Errorf("order %s not found in database", orderNumber)
	}

	// Проверяем статус заказа
	if !w.isOrderActive(order.Status) {
		w.logger.Warn("Order is not in active status",
			zap.String("order_number", orderNumber),
			zap.String("status", string(order.Status)),
		)
		return fmt.Errorf("order %s is not in active status (current: %s)", orderNumber, order.Status)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Проверяем, не отслеживается ли уже этот заказ
	if _, exists := w.tasks[orderNumber]; exists {
		w.logger.Debug("Order is already being tracked",
			zap.String("order_number", orderNumber),
		)
		return fmt.Errorf("order %s is already being tracked", orderNumber)
	}

	// Создаем задачу
	ctxPoll, cancel := context.WithCancel(context.Background())
	task := &orderTask{
		orderNumber: orderNumber,
		cancel:      cancel,
		status:      order.Status,
		lastPolled:  time.Time{},
	}

	w.tasks[orderNumber] = task

	// Запускаем горутину для опроса
	w.wg.Add(1)
	go w.pollOrder(ctxPoll, orderNumber)

	// Отправляем заказ на немедленную обработку
	w.enqueueOrderForProcessing(orderNumber)

	w.logger.Info("Added order to worker",
		zap.String("order_number", orderNumber),
		zap.String("status", string(order.Status)),
		zap.Int("total_tasks", len(w.tasks)),
	)
	return nil
}

// RemoveOrder удаляет заказ из отслеживания
func (w *orderWorker) RemoveOrder(orderNumber string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.removeTaskUnsafe(orderNumber)
}

// removeTaskUnsafe удаляет задачу без блокировки
func (w *orderWorker) removeTaskUnsafe(orderNumber string) error {
	task, exists := w.tasks[orderNumber]
	if !exists {
		w.logger.Debug("Order is not being tracked",
			zap.String("order_number", orderNumber),
		)
		return fmt.Errorf("order %s is not being tracked", orderNumber)
	}

	task.cancel()
	delete(w.tasks, orderNumber)

	w.logger.Info("Stopped tracking order",
		zap.String("order_number", orderNumber),
		zap.String("status", string(task.status)),
		zap.Int("remaining_tasks", len(w.tasks)),
	)
	return nil
}

// AddBatch добавляет несколько заказов для отслеживания
func (w *orderWorker) AddBatch(ctx context.Context, orders []string, interval time.Duration) error {
	w.logger.Info("Adding batch of orders",
		zap.Int("order_count", len(orders)),
		zap.Duration("interval", interval),
	)

	var errors []error
	successCount := 0

	for _, order := range orders {
		if err := w.AddOrder(ctx, order, interval); err != nil {
			errors = append(errors, fmt.Errorf("order %s: %w", order, err))
			w.logger.Warn("Failed to add order in batch",
				zap.String("order_number", order),
				zap.Error(err),
			)
		} else {
			successCount++
		}
	}

	w.logger.Info("Batch add completed",
		zap.Int("successful", successCount),
		zap.Int("failed", len(errors)),
	)

	if len(errors) > 0 {
		return fmt.Errorf("batch add completed with %d errors", len(errors))
	}

	return nil
}

// ==================== Получение информации ====================

// Results возвращает канал с результатами опросов
func (w *orderWorker) Results() <-chan PollResult {
	return w.resultChan
}

// GetActiveOrders возвращает список отслеживаемых заказов
func (w *orderWorker) GetActiveOrders() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	orders := make([]string, 0, len(w.tasks))
	for orderNumber := range w.tasks {
		orders = append(orders, orderNumber)
	}
	return orders
}

// GetOrderStats возвращает статистику по заказам
func (w *orderWorker) GetOrderStats() map[entity.OrderStatus]int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := make(map[entity.OrderStatus]int)
	for _, task := range w.tasks {
		stats[task.status]++
	}

	w.logger.Debug("Order statistics",
		zap.Any("stats", stats),
		zap.Int("total_tasks", len(w.tasks)),
	)

	return stats
}

// ==================== Перезагрузка из БД ====================

// ReloadOrdersFromDB перечитывает активные заказы из базы данных
func (w *orderWorker) ReloadOrdersFromDB(ctx context.Context) error {
	w.logger.Debug("Reloading orders from database")

	activeOrders, err := w.repo.Order().GetActive(ctx)
	if err != nil {
		w.logger.Error("Failed to get active orders from DB", zap.Error(err))
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Создаем map активных заказов для быстрого поиска
	activeMap := make(map[string]*entity.Order)
	for _, order := range activeOrders {
		activeMap[order.Number] = &order
	}

	// Обновляем существующие задачи и добавляем новые
	added := 0
	updated := 0
	for _, order := range activeOrders {
		if w.updateOrCreateTask(&order) {
			added++
		} else {
			updated++
		}
	}

	// Удаляем задачи для заказов, которых больше нет в активных
	removed := w.cleanupStaleTasks(activeMap)

	w.logger.Info("Reloaded orders from DB",
		zap.Int("active_in_db", len(activeOrders)),
		zap.Int("tasks_in_worker", len(w.tasks)),
		zap.Int("added", added),
		zap.Int("updated", updated),
		zap.Int("removed", removed),
	)

	return nil
}

// updateOrCreateTask обновляет существующую задачу или создает новую
// Возвращает true, если задача была создана
func (w *orderWorker) updateOrCreateTask(order *entity.Order) bool {
	if task, exists := w.tasks[order.Number]; exists {
		// Обновляем статус существующей задачи
		if task.status != order.Status {
			w.logger.Debug("Updated order status",
				zap.String("order_number", order.Number),
				zap.String("old_status", string(task.status)),
				zap.String("new_status", string(order.Status)),
			)
			task.status = order.Status
		}
		return false
	} else if w.isOrderActive(order.Status) {
		// Создаем новую задачу только для активных заказов
		ctxPoll, cancel := context.WithCancel(context.Background())
		w.tasks[order.Number] = &orderTask{
			orderNumber: order.Number,
			cancel:      cancel,
			status:      order.Status,
			lastPolled:  time.Time{},
		}

		// Запускаем горутину для опроса
		w.wg.Add(1)
		go w.pollOrder(ctxPoll, order.Number)

		w.logger.Info("Loaded order from DB",
			zap.String("order_number", order.Number),
			zap.String("status", string(order.Status)),
		)
		return true
	}
	return false
}

// cleanupStaleTasks удаляет устаревшие задачи
func (w *orderWorker) cleanupStaleTasks(activeOrders map[string]*entity.Order) int {
	removed := 0
	for orderNumber, task := range w.tasks {
		if _, exists := activeOrders[orderNumber]; !exists {
			task.cancel()
			delete(w.tasks, orderNumber)
			removed++

			w.logger.Info("Removed stale order from worker",
				zap.String("order_number", orderNumber),
				zap.String("status", string(task.status)),
			)
		}
	}
	return removed
}

// CleanupStaleOrders очищает устаревшие задачи (публичный метод)
func (w *orderWorker) CleanupStaleOrders(ctx context.Context) error {
	w.logger.Info("Cleaning up stale orders")
	return w.ReloadOrdersFromDB(ctx)
}

// ==================== Внутренние обработчики ====================

// startOrderHandlers запускает обработчики заказов
func (w *orderWorker) startOrderHandlers(ctx context.Context) {
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.orderHandler(ctx, i)
	}
	w.logger.Debug("Started order handlers", zap.Int("count", w.workers))
}

// startDBScheduler запускает планировщик перечитывания БД
func (w *orderWorker) startDBScheduler(ctx context.Context) {
	w.wg.Add(1)
	go w.dbScheduler(ctx)
	w.logger.Debug("Started DB scheduler")
}

// startOrderDistributor запускает распределитель заказов
func (w *orderWorker) startOrderDistributor(ctx context.Context) {
	w.wg.Add(1)
	go w.orderDistributor(ctx)
	w.logger.Debug("Started order distributor")
}

// orderHandler обрабатывает заказы из канала
func (w *orderWorker) orderHandler(ctx context.Context, workerID int) {
	defer w.wg.Done()

	w.logger.Debug("Order handler started",
		zap.Int("worker_id", workerID),
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("Order handler stopped (context canceled)",
				zap.Int("worker_id", workerID),
			)
			return
		case <-w.stopChan:
			w.logger.Debug("Order handler stopped (worker stopped)",
				zap.Int("worker_id", workerID),
			)
			return
		case orderNumber := <-w.orderChan:
			w.logger.Debug("Processing order",
				zap.Int("worker_id", workerID),
				zap.String("order_number", orderNumber),
			)
			w.processOrder(ctx, orderNumber)
		}
	}
}

// dbScheduler планирует перечитывание заказов из БД
func (w *orderWorker) dbScheduler(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.dbCheckInterval)
	defer ticker.Stop()

	w.logger.Debug("DB scheduler started",
		zap.Duration("interval", w.dbCheckInterval),
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("DB scheduler stopped (context canceled)")
			return
		case <-w.stopChan:
			w.logger.Debug("DB scheduler stopped (worker stopped)")
			return
		case <-ticker.C:
			w.logger.Debug("Scheduled DB reload triggered")
			if err := w.ReloadOrdersFromDB(ctx); err != nil {
				w.logger.Error("Failed to reload orders from DB", zap.Error(err))
			}
		}
	}
}

// orderDistributor распределяет заказы для опроса
func (w *orderWorker) orderDistributor(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Debug("Order distributor started",
		zap.Duration("interval", w.interval),
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("Order distributor stopped (context canceled)")
			return
		case <-w.stopChan:
			w.logger.Debug("Order distributor stopped (worker stopped)")
			return
		case orderNumber := <-w.orderChan:
			w.logger.Debug("Immediate poll requested",
				zap.String("order_number", orderNumber),
			)
			w.pollImmediately(ctx, orderNumber)
		case <-ticker.C:
			w.logger.Debug("Distributing regular polls")
			w.distributeRegularPolls()
		}
	}
}

// ==================== Опрос заказов ====================

// pollOrder периодически опрашивает заказ
func (w *orderWorker) pollOrder(ctx context.Context, orderNumber string) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Debug("Started polling order",
		zap.String("order_number", orderNumber),
		zap.Duration("interval", w.interval),
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("Polling stopped for order (context canceled)",
				zap.String("order_number", orderNumber),
			)
			return
		case <-w.stopChan:
			w.logger.Debug("Polling stopped for order (worker stopped)",
				zap.String("order_number", orderNumber),
			)
			return
		case <-ticker.C:
			w.logger.Debug("Polling order",
				zap.String("order_number", orderNumber),
			)
			w.processOrder(ctx, orderNumber)
			w.checkOrderStatus(ctx, orderNumber)
		}
	}
}

// processOrder обрабатывает один заказ
func (w *orderWorker) processOrder(ctx context.Context, orderNumber string) {
	// Проверяем актуальность заказа
	if !w.isOrderStillActive(ctx, orderNumber) {
		return
	}

	// Выполняем опрос
	orderInfo, err := w.pollExternalAPI(ctx, orderNumber)

	// Обновляем время последнего опроса
	w.updateLastPolled(orderNumber)

	// Отправляем результат
	w.sendPollResult(orderNumber, orderInfo, err)

	// Логируем результат
	if err != nil {
		w.logger.Error("Failed to poll external API",
			zap.String("order_number", orderNumber),
			zap.Error(err),
		)
	} else {
		w.logger.Debug("Successfully polled order",
			zap.String("order_number", orderNumber),
			zap.Any("order_info", orderInfo),
		)
	}
}

// pollImmediately выполняет немедленный опрос заказа
func (w *orderWorker) pollImmediately(ctx context.Context, orderNumber string) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.processOrder(ctx, orderNumber)
	}()
}

// distributeRegularPolls распределяет заказы для регулярного опроса
func (w *orderWorker) distributeRegularPolls() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	now := time.Now()
	pollCount := 0

	for orderNumber, task := range w.tasks {
		if w.shouldPoll(task.status, task.lastPolled, now) {
			w.enqueueOrderForProcessing(orderNumber)
			task.lastPolled = now
			pollCount++
		}
	}

	if pollCount > 0 {
		w.logger.Debug("Distributed regular polls",
			zap.Int("poll_count", pollCount),
			zap.Int("total_tasks", len(w.tasks)),
		)
	}
}

// ==================== Вспомогательные методы ====================

// isOrderActive проверяет активный ли статус у заказа
func (w *orderWorker) isOrderActive(status entity.OrderStatus) bool {
	return status == entity.OrderStatusNew || status == entity.OrderStatusProcessing
}

// isOrderStillActive проверяет актуальность заказа
func (w *orderWorker) isOrderStillActive(ctx context.Context, orderNumber string) bool {
	order, err := w.repo.Order().GetByNumber(ctx, orderNumber)
	if err != nil {
		w.logger.Error("Failed to get order from DB",
			zap.String("order_number", orderNumber),
			zap.Error(err),
		)
		return false
	}

	if order == nil {
		w.logger.Warn("Order not found in DB, removing from worker",
			zap.String("order_number", orderNumber),
		)
		errRemove := w.RemoveOrder(orderNumber)
		if errRemove != nil {
			return false
		}
		return false
	}

	if !w.isOrderActive(order.Status) {
		w.logger.Info("Order is no longer active, removing from worker",
			zap.String("order_number", orderNumber),
			zap.String("status", string(order.Status)),
		)
		errRemove := w.RemoveOrder(orderNumber)
		if errRemove != nil {
			return false
		}
		return false
	}

	return true
}

// checkOrderStatus проверяет статус заказа в БД
func (w *orderWorker) checkOrderStatus(ctx context.Context, orderNumber string) {
	order, err := w.repo.Order().GetByNumber(ctx, orderNumber)
	if err != nil {
		w.logger.Error("Failed to check order status in DB",
			zap.String("order_number", orderNumber),
			zap.Error(err),
		)
		return
	}

	if order == nil || !w.isOrderActive(order.Status) {
		errRemove := w.RemoveOrder(orderNumber)
		if errRemove != nil {
			return
		}
	}
}

// shouldPoll определяет, нужно ли опрашивать заказ
func (w *orderWorker) shouldPoll(status entity.OrderStatus, lastPolled time.Time, now time.Time) bool {
	if !w.isOrderActive(status) {
		return false
	}

	if lastPolled.IsZero() || now.Sub(lastPolled) > w.interval {
		return true
	}

	return false
}

// pollExternalAPI опрашивает внешнее API
func (w *orderWorker) pollExternalAPI(ctx context.Context, orderNumber string) (*orderclient.OrderResponse, error) {
	ctxPoll, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	w.logger.Debug("Polling external API",
		zap.String("order_number", orderNumber),
		zap.Duration("timeout", pollTimeout),
	)

	return w.client.GetOrderInfo(ctxPoll, orderNumber)
}

// updateLastPolled обновляет время последнего опроса
func (w *orderWorker) updateLastPolled(orderNumber string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if task, exists := w.tasks[orderNumber]; exists {
		task.lastPolled = time.Now()
	}
}

// sendPollResult отправляет результат опроса в канал
func (w *orderWorker) sendPollResult(orderNumber string, orderInfo *orderclient.OrderResponse, err error) {
	result := PollResult{
		OrderNumber: orderNumber,
		OrderInfo:   orderInfo,
		Error:       err,
	}

	select {
	case w.resultChan <- result:
		// Результат успешно отправлен
	default:
		w.logger.Warn("Result channel is full, dropping result",
			zap.String("order_number", orderNumber),
			zap.Int("channel_capacity", cap(w.resultChan)),
		)
	}
}

// enqueueOrderForProcessing ставит заказ в очередь на обработку
func (w *orderWorker) enqueueOrderForProcessing(orderNumber string) {
	select {
	case w.orderChan <- orderNumber:
		// Заказ успешно добавлен в очередь
	default:
		w.logger.Warn("Order channel is full, order will be processed later",
			zap.String("order_number", orderNumber),
			zap.Int("channel_capacity", cap(w.orderChan)),
		)
	}
}
