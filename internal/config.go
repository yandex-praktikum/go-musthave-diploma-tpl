package internal

var Config = config{}

type config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
	JwtSecret     string `env:"JWT_SECRET"`
}
