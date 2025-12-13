package worker

import (
	"context"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/repository"
	"log"
	"sync"
	"time"
)

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

type orderWorker struct {
	client          orderclient.Client
	repo            repository.Repository
	workers         int
	interval        time.Duration
	dbCheckInterval time.Duration
	tasks           map[string]*orderTask
	mu              sync.RWMutex
	wg              sync.WaitGroup
	stopChan        chan struct{}
	isRunning       bool
	resultChan      chan PollResult
	orderChan       chan string
}

type orderTask struct {
	orderNumber string
	cancel      context.CancelFunc
	lastPolled  time.Time
	status      entity.OrderStatus
}

// Config конфигурация воркера
type Config struct {
	Client          orderclient.Client
	Repo            repository.Repository
	Workers         int
	Interval        time.Duration
	DBCheckInterval time.Duration
	OrderChanSize   int
	ResultChanSize  int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig(client orderclient.Client, repo repository.Repository) Config {
	return Config{
		Client:          client,
		Repo:            repo,
		Workers:         3,
		Interval:        15 * time.Second,
		DBCheckInterval: 30 * time.Second,
		OrderChanSize:   100,
		ResultChanSize:  100,
	}
}

// NewOrderWorker создает нового воркера для заказов
func NewOrderWorker(config Config) OrderWorker {
	return &orderWorker{
		client:          config.Client,
		repo:            config.Repo,
		workers:         config.Workers,
		interval:        config.Interval,
		dbCheckInterval: config.DBCheckInterval,
		tasks:           make(map[string]*orderTask),
		stopChan:        make(chan struct{}),
		resultChan:      make(chan PollResult, config.ResultChanSize),
		orderChan:       make(chan string, config.OrderChanSize),
	}
}

// Start запускает воркер
func (w *orderWorker) Start(ctx context.Context) error {
	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	w.isRunning = true

	// Загружаем активные заказы из БД при старте
	if err := w.ReloadOrdersFromDB(ctx); err != nil {
		return fmt.Errorf("failed to load orders from DB: %w", err)
	}

	// Запускаем воркеры для обработки заказов
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.processOrders(ctx, i)
	}

	// Запускаем планировщик для перечитывания заказов из БД
	w.wg.Add(1)
	go w.scheduleDBReload(ctx)

	// Запускаем распределитель заказов
	w.wg.Add(1)
	go w.distributeOrders(ctx)

	log.Printf("Order worker started with %d workers", w.workers)
	return nil
}

// Stop останавливает воркер
func (w *orderWorker) Stop() {
	if !w.isRunning {
		return
	}

	close(w.stopChan)
	w.wg.Wait()
	close(w.resultChan)
	w.isRunning = false

	log.Println("Order worker stopped")
}

// ReloadOrdersFromDB перечитывает активные заказы из базы данных
func (w *orderWorker) ReloadOrdersFromDB(ctx context.Context) error {
	activeOrders, err := w.repo.GetActiveOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Обновляем существующие задачи и добавляем новые
	currentOrders := make(map[string]bool)
	for _, order := range activeOrders {
		currentOrders[order.Number] = true

		if task, exists := w.tasks[order.Number]; exists {
			// Обновляем статус существующей задачи
			task.status = order.Status
		} else {
			// Добавляем новую задачу
			ctxPoll, cancel := context.WithCancel(context.Background())
			w.tasks[order.Number] = &orderTask{
				orderNumber: order.Number,
				cancel:      cancel,
				status:      order.Status,
				lastPolled:  time.Time{},
			}

			// Запускаем горутину для опроса этого заказа
			w.wg.Add(1)
			go w.pollOrder(ctxPoll, order.Number)
			log.Printf("Loaded order %s from DB (status: %s)", order.Number, order.Status)
		}
	}

	// Удаляем задачи для заказов, которых больше нет в активных
	for orderNumber, task := range w.tasks {
		if !currentOrders[orderNumber] {
			task.cancel()
			delete(w.tasks, orderNumber)
			log.Printf("Removed order %s from worker (no longer active in DB)", orderNumber)
		}
	}

	log.Printf("Reloaded orders from DB: %d active, %d in worker", len(activeOrders), len(w.tasks))
	return nil
}

// AddOrder добавляет заказ для отслеживания
func (w *orderWorker) AddOrder(ctx context.Context, orderNumber string, interval time.Duration) error {
	// Проверяем существование заказа в БД
	order, err := w.repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to check order in DB: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order %s not found in database", orderNumber)
	}

	// Проверяем, активен ли заказ
	if order.Status != entity.OrderStatusNew && order.Status != entity.OrderStatusProcessing {
		return fmt.Errorf("order %s is not in active status (current: %s)", orderNumber, order.Status)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Проверяем, не отслеживается ли уже этот заказ
	if _, exists := w.tasks[orderNumber]; exists {
		return fmt.Errorf("order %s is already being tracked", orderNumber)
	}

	// Добавляем задачу
	ctxPoll, cancel := context.WithCancel(context.Background())
	w.tasks[orderNumber] = &orderTask{
		orderNumber: orderNumber,
		cancel:      cancel,
		status:      order.Status,
		lastPolled:  time.Time{},
	}

	// Запускаем горутину для опроса этого заказа
	w.wg.Add(1)
	go w.pollOrder(ctxPoll, orderNumber)

	// Отправляем заказ в канал для немедленной обработки
	select {
	case w.orderChan <- orderNumber:
		log.Printf("Added order %s to worker (status: %s)", orderNumber, order.Status)
	default:
		log.Printf("Order channel is full, order %s will be processed later", orderNumber)
	}

	return nil
}

// RemoveOrder удаляет заказ из отслеживания
func (w *orderWorker) RemoveOrder(orderNumber string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	task, exists := w.tasks[orderNumber]
	if !exists {
		return fmt.Errorf("order %s is not being tracked", orderNumber)
	}

	// Отменяем контекст, чтобы остановить опрос
	task.cancel()
	delete(w.tasks, orderNumber)

	log.Printf("Stopped tracking order %s", orderNumber)
	return nil
}

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

// scheduleDBReload планирует перечитывание заказов из БД
func (w *orderWorker) scheduleDBReload(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.dbCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			if err := w.ReloadOrdersFromDB(ctx); err != nil {
				log.Printf("Failed to reload orders from DB: %v", err)
			}
		}
	}
}

// distributeOrders распределяет заказы между воркерами
func (w *orderWorker) distributeOrders(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case orderNumber := <-w.orderChan:
			// Отправляем заказ на обработку (просто для немедленного опроса)
			w.pollImmediately(ctx, orderNumber)
		case <-ticker.C:
			// Распределяем заказы для регулярного опроса
			w.distributeRegularPolls(ctx)
		}
	}
}

// distributeRegularPolls распределяет заказы для регулярного опроса
func (w *orderWorker) distributeRegularPolls(ctx context.Context) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	now := time.Now()
	for orderNumber, task := range w.tasks {
		// Проверяем, нужно ли опрашивать этот заказ
		if shouldPoll(task.status, task.lastPolled, w.interval, now) {
			select {
			case w.orderChan <- orderNumber:
				task.lastPolled = now
			default:
				log.Printf("Order channel is full, skipping order %s", orderNumber)
			}
		}
	}
}

// shouldPoll определяет, нужно ли опрашивать заказ
func shouldPoll(status entity.OrderStatus, lastPolled time.Time, interval time.Duration, now time.Time) bool {
	// Если заказ в финальном статусе, не опрашиваем
	if status == entity.OrderStatusInvalid || status == entity.OrderStatusProcessed {
		return false
	}

	// Если никогда не опрашивали или прошло больше интервала
	if lastPolled.IsZero() || now.Sub(lastPolled) > interval {
		return true
	}

	return false
}

// pollImmediately выполняет немедленный опрос заказа
func (w *orderWorker) pollImmediately(ctx context.Context, orderNumber string) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.doPoll(ctx, orderNumber)
	}()
}

// processOrders обрабатывает заказы
func (w *orderWorker) processOrders(ctx context.Context, workerID int) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case orderNumber := <-w.orderChan:
			w.doPoll(ctx, orderNumber)
		}
	}
}

// pollOrder периодически опрашивает заказ
func (w *orderWorker) pollOrder(ctx context.Context, orderNumber string) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Polling stopped for order %s", orderNumber)
			return
		case <-w.stopChan:
			log.Printf("Worker stopped, exiting poll for order %s", orderNumber)
			return
		case <-ticker.C:
			w.doPoll(ctx, orderNumber)

			// Проверяем статус заказа в БД после опроса
			w.checkOrderStatus(ctx, orderNumber)
		}
	}
}

// checkOrderStatus проверяет статус заказа в БД
func (w *orderWorker) checkOrderStatus(ctx context.Context, orderNumber string) {
	order, err := w.repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		log.Printf("Failed to check order %s status in DB: %v", orderNumber, err)
		return
	}

	if order == nil {
		// Заказ удален из БД, удаляем из воркера
		w.RemoveOrder(orderNumber)
		return
	}

	// Обновляем статус задачи
	w.mu.Lock()
	if task, exists := w.tasks[orderNumber]; exists {
		task.status = order.Status

		// Если заказ перешел в финальный статус, удаляем его
		if order.Status == entity.OrderStatusInvalid || order.Status == entity.OrderStatusProcessed {
			w.mu.Unlock()
			w.RemoveOrder(orderNumber)
			return
		}
	}
	w.mu.Unlock()
}

// doPoll выполняет один опрос заказа
func (w *orderWorker) doPoll(ctx context.Context, orderNumber string) {
	// Сначала проверяем актуальность заказа в БД
	order, err := w.repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		// Если ошибка при получении заказа, логируем и продолжаем
		log.Printf("Failed to get order %s from DB: %v", orderNumber, err)
	} else if order == nil {
		// Заказ удален из БД, удаляем из воркера
		w.RemoveOrder(orderNumber)
		return
	} else if order.Status != entity.OrderStatusNew && order.Status != entity.OrderStatusProcessing {
		// Заказ уже не в активном статусе, удаляем из воркера
		w.RemoveOrder(orderNumber)
		return
	}

	// Выполняем опрос внешнего API
	ctxPoll, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	orderInfo, err := w.client.GetOrderInfo(ctxPoll, orderNumber)

	// Обновляем время последнего опроса
	w.mu.Lock()
	if task, exists := w.tasks[orderNumber]; exists {
		task.lastPolled = time.Now()
	}
	w.mu.Unlock()

	// Отправляем результат в канал
	select {
	case w.resultChan <- PollResult{
		OrderNumber: orderNumber,
		OrderInfo:   orderInfo,
		Error:       err,
	}:
		// Результат успешно отправлен
	default:
		log.Printf("Result channel is full, dropping result for order %s", orderNumber)
	}
}

// AddBatch добавляет несколько заказов для отслеживания
func (w *orderWorker) AddBatch(ctx context.Context, orders []string, interval time.Duration) error {
	for _, order := range orders {
		if err := w.AddOrder(ctx, order, interval); err != nil {
			// Логируем ошибку, но продолжаем добавлять остальные заказы
			log.Printf("Failed to add order %s: %v", order, err)
		}
	}
	return nil
}

// GetOrderStats возвращает статистику по заказам
func (w *orderWorker) GetOrderStats() map[entity.OrderStatus]int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := make(map[entity.OrderStatus]int)
	for _, task := range w.tasks {
		stats[task.status]++
	}
	return stats
}

// CleanupStaleOrders очищает устаревшие задачи
func (w *orderWorker) CleanupStaleOrders(ctx context.Context) error {
	activeOrders, err := w.repo.GetActiveOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active orders: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Создаем map активных заказов для быстрого поиска
	activeMap := make(map[string]bool)
	for _, order := range activeOrders {
		activeMap[order.Number] = true
	}

	// Удаляем задачи для заказов, которых нет в активных
	for orderNumber, task := range w.tasks {
		if !activeMap[orderNumber] {
			task.cancel()
			delete(w.tasks, orderNumber)
			log.Printf("Cleaned up stale order %s from worker", orderNumber)
		}
	}

	return nil
}
