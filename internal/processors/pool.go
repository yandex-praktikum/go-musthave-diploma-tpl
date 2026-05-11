package processors

import (
	"context"
	"sync"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
	"github.com/KalessinD/gophermart/internal/repositories/file"
	"go.uber.org/zap"
)

type (
	// Объединение сотрудников в коллектив для обработки заказов сообща
	WorkerPool struct {
		workers      []WorkerInterface // open-space для сотрдуников
		wg           sync.WaitGroup
		mx           sync.Mutex
		workerCh     chan *Task         // для передачи заданий сотрудникам
		taskCh       <-chan *Task       // входной канал для получения заданий от сервиса
		pauseCh      chan time.Duration // входной канал для диспетчера, чтобы объявить всеобщий перерыв
		cancel       context.CancelFunc
		log          *zap.Logger
		pendingTasks TaskList      // очередь LIFO на случай, если сотрудникам в спринт новые задачи не получается добавить
		hasTasks     chan struct{} // канал-будильник для диспетчера, если тот задремал без работы
		dumper       file.PersistStorageInterface
	}

	// Интерфейс для управления коллективом сотрудников
	WorkerPoolInterface interface {
		RestoreQueue()
		Start(ctx context.Context)
		Stop()
		Wait()
	}
)

// Метод создания рабочего коллектива
func NewQueueProcessor(
	poolSize,
	bufSize int,
	log *zap.Logger,
	inCh <-chan *Task,
	pCh chan time.Duration,
	dumper file.PersistStorageInterface,
	action TaskProcessor,
) (WorkerPoolInterface, error) {
	pool := &WorkerPool{
		workers:      make([]WorkerInterface, poolSize),
		wg:           sync.WaitGroup{},
		mx:           sync.Mutex{},
		taskCh:       inCh,
		pauseCh:      pCh,
		workerCh:     make(chan *Task, bufSize),
		pendingTasks: make(TaskList, 0, bufSize),
		hasTasks:     make(chan struct{}, 1),
		log:          log,
		dumper:       dumper,
	}

	for i := range poolSize {
		worker, err := NewWorker(pool.workerCh, log, action)
		if err != nil {
			return nil, err
		}
		pool.workers[i] = worker
	}

	return pool, nil
}

// Обработка заданий из единого окна заявок
func (wp *WorkerPool) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// При Graceful Shutdown стек заданий спасаем в диспетчере
			wp.Stop()
			return
		case task, ok := <-wp.taskCh:
			if !ok {
				return
			}

			wp.mx.Lock()
			wp.pendingTasks = append(wp.pendingTasks, task)
			wp.mx.Unlock()

			select {
			case wp.hasTasks <- struct{}{}:
				// толкнули диспетчера, чтобы не спал
			default:
				// если диспетчер занят, то просто идём по своим делам и никого не держим - задание в очереди
			}
		}
	}
}

// Просим рабочий коллектив приступить к работе
func (wp *WorkerPool) Start(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	wp.cancel = cancel

	wp.wg.Add(len(wp.workers) + 2) // + dispatcher + inputQueueReader

	for _, worker := range wp.workers {
		go func(ctx context.Context, worker WorkerInterface, log *zap.Logger) {
			defer wp.wg.Done()
			worker.Run(ctx, wp.pauseCh)
			log.Sugar().Infof("worker %s is done", worker.ID())
		}(ctx, worker, wp.log)
	}

	// inputQueueReader
	go func() {
		defer wp.wg.Done()
		wp.run(ctx)
	}()

	// dispatcher
	go func() {
		defer wp.wg.Done()
		wp.runDispatcher(ctx)
	}()

	// waiter is waiting to do some durty works at finish
	go func() {
		wp.wg.Wait()
		close(wp.workerCh)
	}()
}

// Отправляем коллектив сотрудников по домам
func (wp *WorkerPool) Stop() {
	wp.mx.Lock()
	defer wp.mx.Unlock()
	if wp.cancel != nil {
		wp.cancel()
		wp.cancel = nil
	}
}

// Ожидаем когда сотрудники доделают работу
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool) runDispatcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// При Graceful Shutdown надо сделать дамп очереди в локаьный файл
			wp.dumpQueue()
			return
		case delay, opened := <-wp.pauseCh:
			if !opened {
				return
			}
			// пора отдохнуть, перестаём на заданное время выдавать задания в работу
			wp.sleepForAWhile(ctx, delay)

			continue

		case <-wp.hasTasks:
			// есть для нас работёнка
		}

		for {
			wp.mx.Lock()

			if len(wp.pendingTasks) == 0 {
				wp.mx.Unlock()
				break
			}

			// LIFO
			lastIdx := len(wp.pendingTasks) - 1
			task := wp.pendingTasks[lastIdx]

			// зануляем ссылку, чтобы GC мог удалить объект
			wp.pendingTasks[lastIdx] = nil

			// Урезаем слайс. Теперь последний элемент выпал, но capacity осталась той же.
			wp.pendingTasks = wp.pendingTasks[:lastIdx]

			wp.mx.Unlock()

			select {
			case wp.workerCh <- task:

			case delay, opened := <-wp.pauseCh:
				if !opened {
					return
				}

				// вернём задачу в стек, чтобы не потерялть её
				wp.mx.Lock()
				wp.pendingTasks = append(wp.pendingTasks, task)
				wp.mx.Unlock()

				// пора отдохнуть, перестаём на заданное время выдавать задания в работу
				wp.sleepForAWhile(ctx, delay)

			case <-ctx.Done():
				// При Graceful Shutdown надо сделать дамп очереди в локальный файл

				// вернём задачу в стек, чтобы не потерялть её
				wp.mx.Lock()
				wp.pendingTasks = append(wp.pendingTasks, task)
				wp.mx.Unlock()

				wp.dumpQueue()
				return
			}
		}
	}
}

func (wp *WorkerPool) sleepForAWhile(ctx context.Context, delay time.Duration) {
	ticker := time.NewTicker(delay)
	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		// перерыв окончен
	}
}

func (wp *WorkerPool) dumpQueue() {
	wp.log.Info("goinfg to dump order's queue")

	orders := make(models.OrdersList, 0, len(wp.pendingTasks))

	for _, task := range wp.pendingTasks {
		order := models.Order(*task)
		orders = append(orders, &order)
	}

	queueIsEmpty := false

	// вычитаем всё из рабочей очереди назад
	for !queueIsEmpty {
		select {
		case task, ok := <-wp.workerCh:
			if !ok {
				break
			}
			order := models.Order(*task)
			orders = append(orders, &order)
		default:
			queueIsEmpty = true
		}
	}

	err := wp.dumper.Save(orders)
	if err != nil {
		wp.log.Error("can't dump order's queue", zap.Error(err))
	} else {
		wp.log.Info("order's queue has been dumped successfully")
	}
}

func (wp *WorkerPool) RestoreQueue() {
	// защита от умника, в целом конкуренции тут быть не должно
	wp.mx.Lock()
	defer wp.mx.Unlock()

	orders, err := wp.dumper.Restore()
	if err != nil {
		wp.log.Error("can't restore order's queue from dump", zap.Error(err))
		return
	}

	err = wp.dumper.Erase()
	if err != nil {
		wp.log.Error("queue won't be restored due to errro while ersaing dump, old dump is preserved", zap.Error(err))
		return
	}

	for _, order := range orders {
		task := Task(*order)
		wp.pendingTasks = append(wp.pendingTasks, &task)
	}

	if len(orders) > 0 {
		// скажем диспетчеру, что пришло время поработать
		wp.hasTasks <- struct{}{}
	}

	wp.log.Info("order's queue has been restored from dump successfully")
}
