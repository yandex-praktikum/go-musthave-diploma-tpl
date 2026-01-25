package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"musthave/internal/accrual"
	cfg "musthave/internal/config/app"
	"musthave/internal/config/db"
	"musthave/internal/model"
	"musthave/internal/repository"
	"sync"
	"time"
)

type Market struct {
	Lg     *slog.Logger
	Ctx    context.Context
	conn   *sql.DB
	Repo   *repository.Repo
	client *accrual.Client

	HTTPPort    string
	PathStorage string
	DNS         string
	UserCH      map[string]*model.User

	Mu sync.RWMutex

	paramGetStatus time.Duration
}

// Create - заполнение структуры приложения
func Create(ctx context.Context, lg *slog.Logger, cfg *cfg.Config, cl *accrual.Client) (*Market, error) {
	userCH := map[string]*model.User{}
	m := &Market{
		Lg:       lg,
		Ctx:      ctx,
		UserCH:   userCH,
		HTTPPort: cfg.Port,
		client:   cl,
	}

	m.paramGetStatus = time.Second * time.Duration(cfg.ParamGetStatus)
	if cfg.DNS != "" {
		conn, err := db.NewConnection(ctx, cfg.DNS)
		if err != nil {
			lg.Error("db: не удалось подключиться к DB")
			return nil, err
		}
		m.DNS = cfg.DNS
		lg.Info("db: успешно подключились к DB")
		m.conn = conn
		m.Repo = repository.NewRepository(m.conn)
	}

	err := m.init(ctx, lg)
	if err != nil {
		return nil, err
	}
	m.checkUnfinishedOrder(ctx)

	return m, nil
}

// init - инициализация данных
func (m *Market) init(ctx context.Context, lg *slog.Logger) error {
	lg.Info("init.start - инициализация пользователей ")
	err := m.LoadStorageUser(ctx)
	if err != nil {
		lg.Error("init.error - ошибка загрузки пользователей: " + err.Error())
		return err
	}
	lg.Info("init.finish - пользователи успешно подгружены")
	return nil
}

// checkUnfinishedOrder - проверка статуса заказов
func (m *Market) checkUnfinishedOrder(ctx context.Context) {
	m.Lg.Info("checkUnfinishedOrder.Start - старт проверки незавершенных поокупок")
	for _, user := range m.UserCH {
		for _, order := range user.OrderList {
			if order.Status != model.PROCESSED {
				m.Lg.Info(fmt.Sprintf("checkUnfinishedOrder.search - у пользователя: %v - найден заказ с ID: %v, в статусе - %s", user.Login, order.OrderID, order.Status))
				go m.processGetStatus(ctx, user.Login, order.OrderID, m.paramGetStatus)
			}
		}
	}
	m.Lg.Info("checkUnfinishedOrder.finish - список незавершенных поокупок пуст")
}

// processGetStatus - процес получения статуса и баллов по заказу
func (m *Market) processGetStatus(ctx context.Context, login string, orderID int, param time.Duration) {
	processContext, cancel := context.WithCancel(ctx)
	m.Lg.Info("processGetStatus.start - старт процесса опроса статуса по заказу: " + fmt.Sprintf("%v", orderID))
	ticker := time.NewTicker(time.Second * 1) // сразу запускаем процесс
	for {
		m.Lg.Info("processGetStatus.wait - ожидание новой итерации опроса статуса по заказу: " + fmt.Sprintf("%v", orderID))
		select {
		case <-ctx.Done():
			m.Lg.Info("processGetStatus.cancel - контекст процесса был завршен: " + fmt.Sprintf("%v", orderID))
			ticker.Stop()
			cancel()
			return
		case <-ticker.C:
			ticker.Reset(param)
			res, err := m.client.GetAccrual(processContext, m.Lg, orderID)
			if err != nil {
				m.Lg.Error("processGetStatus.error - ошибка получения статуса по заказу: " + fmt.Sprintf("%v", orderID) + ", ошибка: " + err.Error())
				err := m.Repo.SetStatus(ctx, orderID, model.PROCESSING)
				if err != nil {
					m.Lg.Error(fmt.Sprintf("processGetStatus.error - не удалось установить статус %v, для заказа %v, из за ошибки: %v", model.PROCESSING, orderID, err.Error()))
					continue
				}
				continue
			}

			finish, err := m.checkStatus(processContext, res, orderID, login)
			if err != nil {
				m.Lg.Error(fmt.Sprintf("processGetStatus.error - ошибка установки статуса по заказу %v: %v", orderID, err.Error()))
				continue
			}
			m.Mu.RLock()
			order := m.UserCH[login].OrderList[orderID]
			m.Mu.RUnlock()
			if finish {
				m.Lg.Info("processGetStatus.finish - заказ" + res.Order + " имеет финишный статус: " + res.Status + ", процес опроса будет завершен.")
				order.Status = res.Status
				order.Accural = fmt.Sprintf("%v", res.Accrual)
				cancel()
			}
			order.Status = res.Status
			m.Lg.Info("processGetStatus.process - заказ" + res.Order + " не имеет финишный статус: " + res.Status + ", процес опроса будет продолжаться.")
		}
	}
}
