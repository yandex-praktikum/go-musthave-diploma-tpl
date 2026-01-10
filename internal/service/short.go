package service

import (
	"context"
	"database/sql"
	cfg "musthave/internal/config/app"
	"musthave/internal/config/db"
	"musthave/internal/model"
	"musthave/internal/repository"
	"sync"

	"github.com/sirupsen/logrus"
)

const defaultParamDelete = 20 // дефолтный параметр на запуска процесса уадаления

type Short struct {
	Lg   *logrus.Logger
	Ctx  context.Context
	conn *sql.DB
	Repo *repository.Repo

	HTTPPort    string
	PathStorage string
	DNS         string
	UserCH      map[string]*model.User

	mu sync.RWMutex

	//paramDelete time.Duration
}

// NewShort - заполнение структуры приложения
func Create(ctx context.Context, lg *logrus.Logger, cfg *cfg.Config) (*Short, error) {
	userCH := map[string]*model.User{}

	sh := &Short{
		Lg:       lg,
		Ctx:      ctx,
		UserCH:   userCH,
		HTTPPort: cfg.Port,
	}

	if cfg.DNS != "" {
		conn, err := db.NewConnection(cfg.DNS)
		if err != nil {
			lg.Error("db: не удалось подключилиться к DB")
			return nil, err
		}
		sh.DNS = cfg.DNS
		lg.Info("db: успешно подключились к DB")
		sh.conn = conn
		sh.Repo = repository.NewRepository(sh.conn)
	}

	//sh.paramDelete = defaultParamDelete * time.Second
	//if cfg.ParamDelete != 0 {
	//	sh.paramDelete = time.Duration(cfg.ParamDelete) * time.Second
	//}
	//sh.Address = cfg.Address
	//go sh.StartCleanup(ctx, sh.paramDelete)
	//err := sh.LoadStorageURL() //подгрузка кеша
	//if err != nil {
	//	return nil, err
	//}
	return sh, nil
}

// LoadStorageURL - подгрузка в кеш
func (s *Short) LoadStorageURL() error {

	return nil

}

//func (s *Short) StartCleanup(ctx context.Context, param time.Duration) {
//	ticker := time.NewTicker(param)
//	s.Logger.Info("StartCleanup.start - старт процесса очистки помеченных на удаления URL")
//	for {
//		s.Logger.Info("StartCleanup.wait - ожидание новой итерации очистки")
//		select {
//		case <-ctx.Done():
//			s.Logger.Info("StartCleanup.cancel - контекст процесса был завршен")
//			ticker.Stop()
//			return
//		case <-ticker.C:
//			ticker.Reset(param)
//			err := s.DeleteMarkedURLs(ctx)
//			if err != nil {
//				s.Logger.Error("Ошибка при удалении помеченных URL: " + err.Error())
//				continue
//			}
//			s.Logger.Info("StartCleanup.complete - успешная очистка помеченных на удаления URL")
//		}
//	}
//}
