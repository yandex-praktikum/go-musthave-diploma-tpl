package worker

import (
	"context"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
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
}

// OrderAPIClient интерфейс для клиента API заказов
type OrderAPIClient interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*orderclient.OrderResponse, error)
}

type orderWorker struct {
	client      orderclient.Client
	workers     int
	interval    time.Duration
	tasks       map[string]context.CancelFunc
	mu          sync.RWMutex
	wg          sync.WaitGroup
	stopChan    chan struct{}
	isRunning   bool
	resultChan  chan PollResult
	orderBuffer chan string
}

// Config конфигурация воркера
type Config struct {
	Client        orderclient.Client
	Workers       int
	Interval      time.Duration
	BufferSize    int
	ResultBufSize int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig(client orderclient.Client) Config {
	return Config{
		Client:        client,
		Workers:       3,
		Interval:      15 * time.Second,
		BufferSize:    100,
		ResultBufSize: 100,
	}
}

// NewOrderWorker создает нового воркера для заказов
func NewOrderWorker(config Config) OrderWorker {
	return &orderWorker{
		client:      config.Client,
		workers:     config.Workers,
		interval:    config.Interval,
		tasks:       make(map[string]context.CancelFunc),
		stopChan:    make(chan struct{}),
		resultChan:  make(chan PollResult, config.ResultBufSize),
		orderBuffer: make(chan string, config.BufferSize),
	}
}

// Start запускает воркер
func (w *orderWorker) Start(ctx context.Context) error {
	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	w.isRunning = true

	// Запускаем воркеры для обработки буфера заказов
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.processOrderBuffer(ctx, i)
	}

	// Запускаем воркер для чтения из буфера и запуска опросов
	w.wg.Add(1)
	go w.bufferProcessor(ctx)

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

// AddOrder добавляет заказ для отслеживания
func (w *orderWorker) AddOrder(ctx context.Context, orderNumber string, interval time.Duration) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Проверяем, не отслеживается ли уже этот заказ
	if _, exists := w.tasks[orderNumber]; exists {
		return fmt.Errorf("order %s is already being tracked", orderNumber)
	}

	if interval == 0 {
		interval = w.interval
	}

	// Создаем контекст для отмены опроса этого заказа
	ctxPoll, cancel := context.WithCancel(context.Background())
	w.tasks[orderNumber] = cancel

	// Запускаем горутину для периодического опроса заказа
	w.wg.Add(1)
	go w.pollOrder(ctxPoll, orderNumber, interval)

	log.Printf("Started tracking order %s (interval: %v)", orderNumber, interval)
	return nil
}

// RemoveOrder удаляет заказ из отслеживания
func (w *orderWorker) RemoveOrder(orderNumber string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cancel, exists := w.tasks[orderNumber]
	if !exists {
		return fmt.Errorf("order %s is not being tracked", orderNumber)
	}

	// Отменяем контекст, чтобы остановить опрос
	cancel()
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
	for order := range w.tasks {
		orders = append(orders, order)
	}
	return orders
}

// bufferProcessor обрабатывает буфер заказов
func (w *orderWorker) bufferProcessor(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case orderNumber := <-w.orderBuffer:
			// Помещаем заказ в буфер для обработки
			w.wg.Add(1)

			go w.processSingleOrder(ctx, orderNumber)
		}
	}
}

// processOrderBuffer обрабатывает заказы из буфера
func (w *orderWorker) processOrderBuffer(ctx context.Context, workerID int) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval / time.Duration(w.workers))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			// Получаем список заказов для обработки
			orders := w.GetActiveOrders()
			// Распределяем заказы между воркерами
			if len(orders) > 0 {
				// Простой round-robin распределитель
				index := workerID % len(orders)
				if index < len(orders) {
					orderNumber := orders[index]
					select {
					case w.orderBuffer <- orderNumber:
						// Заказ добавлен в буфер
					default:
						log.Printf("Order buffer is full, dropping order %s", orderNumber)
					}
				}
			}
		}
	}
}

// pollOrder периодически опрашивает заказ
func (w *orderWorker) pollOrder(ctx context.Context, orderNumber string, interval time.Duration) {
	defer w.wg.Done()

	ticker := time.NewTicker(interval)
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
			// Опрашиваем заказ
			w.doPoll(ctx, orderNumber)
		}
	}
}

// processSingleOrder обрабатывает один заказ
func (w *orderWorker) processSingleOrder(ctx context.Context, orderNumber string) {
	defer w.wg.Done()

	w.doPoll(ctx, orderNumber)
}

// doPoll выполняет один опрос заказа
func (w *orderWorker) doPoll(ctx context.Context, orderNumber string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	orderInfo, err := w.client.GetOrderInfo(ctx, orderNumber)
	fmt.Println(orderInfo, orderNumber)
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

// CleanupCompletedOrders очищает завершенные заказы
func (w *orderWorker) CleanupCompletedOrders(activeOrders []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Создаем map для быстрого поиска активных заказов
	activeMap := make(map[string]bool)
	for _, order := range activeOrders {
		activeMap[order] = true
	}

	// Удаляем заказы, которых нет в активных
	for orderNumber := range w.tasks {
		if !activeMap[orderNumber] {
			if cancel, exists := w.tasks[orderNumber]; exists {
				cancel()
				delete(w.tasks, orderNumber)
				log.Printf("Cleaned up completed order %s from worker", orderNumber)
			}
		}
	}
}
