package db

import (
	"time"
)

type Config struct {
	Dsn             string
	MaxIdleConn     int
	MaxOpenConn     int
	MaxLifetimeConn time.Duration
}
