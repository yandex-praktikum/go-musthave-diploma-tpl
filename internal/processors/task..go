package processors

import (
	"context"
	"time"

	"github.com/KalessinD/gophermart/internal/models"
)

type (
	Task     models.Order
	TaskList []*Task

	TaskProcessor func(context.Context, chan time.Duration, *Task) error
)
