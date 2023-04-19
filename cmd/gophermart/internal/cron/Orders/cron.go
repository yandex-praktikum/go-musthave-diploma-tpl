package Orders

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/service"
	"github.com/robfig/cron/v3"
	"golang.org/x/exp/slog"
)

type OrderCronJob struct {
	log     *slog.Logger
	service *service.Service
}

func (cj *OrderCronJob) Run() {
	cj.log.Info("Start ordersCronJob cron job")
	cj.service.UpdateOrders(context.Background())
	cj.log.Info("cron job completed successfully")
}

func NewAOrderCronJob(log *slog.Logger, service *service.Service) cron.Job {
	return &OrderCronJob{
		log:     log,
		service: service,
	}
}
