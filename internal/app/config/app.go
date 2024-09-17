package config

type App struct {
	Listen        string `env:"RUN_ADDRESS"`
	AccrualListen string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DBDsn         string `env:"DATABASE_URI"`
}
