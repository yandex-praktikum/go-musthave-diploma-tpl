package cron

import (
	"context"

	"github.com/RedWood011/cmd/gophermart/internal/cron/Orders"
	"github.com/RedWood011/cmd/gophermart/internal/service"
	"github.com/robfig/cron/v3"
	"golang.org/x/exp/slog"
)

type Cron struct {
	cron *cron.Cron
}

func NewCron(
	service *service.Service,
	logger *slog.Logger,
) (*Cron, error) {
	c := cron.New()

	_, err := c.AddJob("@every 3s", Orders.NewAOrderCronJob(logger, service))
	if err != nil {
		return nil, err
	}

	return &Cron{
		cron: c,
	}, nil
}

func (c *Cron) Run(ctx context.Context) {
	c.cron.Start()

	<-ctx.Done()

	<-c.cron.Stop().Done()
}
