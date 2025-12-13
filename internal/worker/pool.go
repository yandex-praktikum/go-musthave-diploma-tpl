package worker

/*
import (
	"context"
	"fmt"
	"github.com/skiphead/go-musthave-diploma-tpl/internal/domain/entity"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skiphead/go-musthave-diploma-tpl/infra/client/orderclient"
)

// Pool представляет воркер-пул для периодического опроса заказов
type Pool struct {
	client     *orderclient.Client
	db         *pgxpool.Pool
	workers    int
	interval   time.Duration
	tasks      map[string]context.CancelFunc
	mu         sync.RWMutex
	wg         sync.WaitGroup
	stopChan   chan struct{}
	isRunning  bool
	resultChan chan Result
}

// Result представляет результат опроса
type Result struct {
	OrderNumber string
	OrderInfo   *orderclient.OrderResponse
	Error       error
}

// Config конфигурация пула
type Config1 struct {
	Client   *orderclient.Client
	DB       *pgxpool.Pool
	Workers  int
	Interval time.Duration
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig1(client *orderclient.Client, db *pgxpool.Pool) Config {
	return Config1{
		Client:   client,
		DB:       db,
		Workers:  3,
		Interval: 15 * time.Second,
	}
}

// NewPool создает новый пул
func NewPool(config Config1) *Pool {
	return &Pool{
		client:     config.Client,
		db:         config.DB,
		workers:    config.Workers,
		interval:   config.Interval,
		tasks:      make(map[string]context.CancelFunc),
		stopChan:   make(chan struct{}),
		resultChan: make(chan Result, 100),
	}
}

// Start запускает пул и загружает активные заказы из БД
func (p *Pool) Start(ctx context.Context) error {
	if p.isRunning {
		return fmt.Errorf("pool is already running")
	}

	p.isRunning = true

	// Загружаем активные заказы из БД
	if err := p.loadActiveOrders(ctx); err != nil {
		return fmt.Errorf("failed to load active orders: %w", err)
	}

	// Запускаем воркеры
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	log.Printf("Pool started with %d workers", p.workers)
	return nil
}

// loadActiveOrders загружает активные заказы из базы данных
func (p *Pool) loadActiveOrders(ctx context.Context) error {
	query := `
		SELECT number
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(ctx, query, entity.OrderStatusNew, entity.OrderStatusProcessing)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var orders []string
	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
		orders = append(orders, number)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	// Добавляем загруженные заказы в пул
	for _, order := range orders {
		// Используем внутренний метод, чтобы не проверять существование в БД
		p.mu.Lock()
		if _, exists := p.tasks[order]; !exists {
			ctx, cancel := context.WithCancel(context.Background())
			p.tasks[order] = cancel
			p.wg.Add(1)
			go p.pollOrder(ctx, order, p.interval)
			log.Printf("Loaded order %s from database", order)
		}
		p.mu.Unlock()
	}

	log.Printf("Loaded %d active orders from database", len(orders))
	return nil
}

// Stop останавливает пул
func (p *Pool) Stop() {
	if !p.isRunning {
		return
	}

	close(p.stopChan)
	p.wg.Wait()
	close(p.resultChan)
	p.isRunning = false

	log.Println("Pool stopped")
}

// Add добавляет заказ для периодического опроса
func (p *Pool) Add(ctx context.Context, orderNumber string, interval time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Проверяем в памяти
	if _, exists := p.tasks[orderNumber]; exists {
		return fmt.Errorf("order %s already in pool", orderNumber)
	}

	// Проверяем в базе данных
	var exists bool
	err := p.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1 AND status IN ($2, $3))",
		orderNumber, entity.OrderStatusNew, entity.OrderStatusProcessing,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("database check failed: %w", err)
	}

	if !exists {
		return fmt.Errorf("order %s not found in database or not in active status", orderNumber)
	}

	if interval == 0 {
		interval = p.interval
	}

	ctxPoll, cancel := context.WithCancel(context.Background())
	p.tasks[orderNumber] = cancel

	// Запускаем горутину для опроса этого заказа
	p.wg.Add(1)
	go p.pollOrder(ctxPoll, orderNumber, interval)

	log.Printf("Added order %s (interval: %v)", orderNumber, interval)
	return nil
}

// Remove удаляет заказ из опроса
func (p *Pool) Remove(orderNumber string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cancel, exists := p.tasks[orderNumber]
	if !exists {
		return fmt.Errorf("order %s not found in pool", orderNumber)
	}

	cancel()
	delete(p.tasks, orderNumber)

	log.Printf("Removed order %s from pool", orderNumber)
	return nil
}

// Results возвращает канал с результатами
func (p *Pool) Results() <-chan Result {
	return p.resultChan
}

// worker воркер (в данной реализации используется для балансировки,
// хотя каждый заказ опрашивается в своей горутине)
func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	select {
	case <-ctx.Done():
		return
	case <-p.stopChan:
		return
	}
}

// pollOrder периодически опрашивает заказ
func (p *Pool) pollOrder(ctx context.Context, orderNumber string, interval time.Duration) {
	defer p.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Polling stopped for order %s", orderNumber)
			return
		case <-p.stopChan:
			log.Printf("Pool stopped, exiting poll for order %s", orderNumber)
			return
		case <-ticker.C:
			p.doPoll(ctx, orderNumber)
		}
	}
}

// doPoll выполняет один опрос заказа
func (p *Pool) doPoll(ctx context.Context, orderNumber string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	orderInfo, err := p.client.GetOrderInfo(ctx, orderNumber)

	// Обновляем статус в базе данных при получении результата
	if orderInfo != nil && orderInfo.Status != "" {
		if err := p.updateOrderStatus(ctx, orderNumber, orderInfo); err != nil {
			log.Printf("Failed to update order %s status: %v", orderNumber, err)
		}
	}

	// Если заказ завершен, удаляем его из пула
	if orderInfo != nil && isFinalStatus(string(orderInfo.Status)) {
		p.Remove(orderNumber)
	}

	select {
	case p.resultChan <- Result{
		OrderNumber: orderNumber,
		OrderInfo:   orderInfo,
		Error:       err,
	}:
	default:
		log.Printf("Result channel is full, dropping result for order %s", orderNumber)
	}
}

// updateOrderStatus обновляет статус заказа в базе данных
func (p *Pool) updateOrderStatus(ctx context.Context, orderNumber string, orderInfo *orderclient.OrderResponse) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE orders
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE number = $3
		RETURNING id
	`

	var id int
	err = tx.QueryRow(ctx, query,
		orderInfo.Status,
		orderInfo.Accrual,
		orderNumber,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("update order failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	log.Printf("Order %s updated with status %s", orderNumber, orderInfo.Status)
	return nil
}

// isFinalStatus проверяет, является ли статус финальным
func isFinalStatus(status string) bool {
	finalStatuses := []string{
		string(entity.OrderStatusInvalid),
		string(entity.OrderStatusProcessed),
	}

	for _, s := range finalStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// AddBatch добавляет несколько заказов
func (p *Pool) AddBatch(ctx context.Context, orders []string, interval time.Duration) error {
	for _, order := range orders {
		if err := p.Add(ctx, order, interval); err != nil {
			// Логируем ошибку, но продолжаем добавлять остальные
			log.Printf("Failed to add order %s: %v", order, err)
		}
	}
	return nil
}

// GetOrders возвращает список отслеживаемых заказов
func (p *Pool) GetOrders() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	orders := make([]string, 0, len(p.tasks))
	for order := range p.tasks {
		orders = append(orders, order)
	}
	return orders
}

// GetActiveOrdersFromDB возвращает активные заказы из базы данных
func (p *Pool) GetActiveOrdersFromDB(ctx context.Context) ([]string, error) {
	query := `
		SELECT number
		FROM orders
		WHERE status IN ($1, $2)
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(ctx, query, entity.OrderStatusNew, entity.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var orders []string
	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		orders = append(orders, number)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// CleanupCompletedOrders удаляет завершенные заказы из пула
func (p *Pool) CleanupCompletedOrders(ctx context.Context) error {
	activeOrders, err := p.GetActiveOrdersFromDB(ctx)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Находим заказы, которые есть в пуле, но нет в активных в БД
	for orderNumber := range p.tasks {
		found := false
		for _, activeOrder := range activeOrders {
			if activeOrder == orderNumber {
				found = true
				break
			}
		}

		if !found {
			// Заказ завершен, удаляем из пула
			if cancel, exists := p.tasks[orderNumber]; exists {
				cancel()
				delete(p.tasks, orderNumber)
				log.Printf("Cleaned up completed order %s from pool", orderNumber)
			}
		}
	}

	return nil
}

*/
