package config

import (
	"errors"
	"flag"
	"time"
)

const (
	DefaultListenAddr               string        = ":9081"
	DefaultProcessingTimeout        time.Duration = 60 * time.Second
	DefaultReadTimeout              time.Duration = 5 * time.Second
	DefaultReadHeaderTimeout        time.Duration = 5 * time.Second
	DefaultWriteTimeout             time.Duration = 10 * time.Second
	DefaultIdleTimeout              time.Duration = 30 * time.Second
	DefaultGracefullShutdownTimeout time.Duration = 5 * time.Second
	DefaultPsqlDSN                  string        = ""
	DefaultAccrualAddress           string        = ""
	DefaultQueueWorkers             int           = 4
	DefaultQueueBufSize             int           = 100
	DefaultWorkerPoolChanBuffer     int           = 32
	DefaultAccrualClientTImeout     time.Duration = 3 * time.Second
	DefaultDumperStoragePath        string        = "/tmp/practicum-gophermart-storage.dumps"

	// got by using `openssl rand -hex 32`
	DefaultServerEncryptionKey string = "c7f7b4036a3fb58734412433cb7a2ed8dec913c650ef8475f05f5b36422cc18d"
)

type (
	GophermartConfig struct {
		ListenAddr               string
		ProcessingTimeout        time.Duration
		ReadTimeout              time.Duration
		ReadHeaderTimeout        time.Duration
		WriteTimeout             time.Duration
		IdleTimeout              time.Duration
		GracefullShutdownTimeout time.Duration
		PsqlDSN                  string
		EncryptionKey            string
		AccrualAddress           string
		QueueBufSize             int
		QueueWorkers             int
		AccrualClientTImeout     time.Duration
		WorkerPoolChanBuffer     int
		DumperStoragePath        string
	}

	GophermartConfigInterface interface {
		UpdateFromEnvironment() error
		UpdateFromCLIArgs() error
		Validate() error
	}
)

func GetDefaultGophermartConfig() *GophermartConfig {
	return &GophermartConfig{
		ListenAddr:               DefaultListenAddr,
		ProcessingTimeout:        DefaultProcessingTimeout,
		ReadTimeout:              DefaultReadTimeout,
		ReadHeaderTimeout:        DefaultReadHeaderTimeout,
		WriteTimeout:             DefaultWriteTimeout,
		IdleTimeout:              DefaultIdleTimeout,
		GracefullShutdownTimeout: DefaultGracefullShutdownTimeout,
		PsqlDSN:                  DefaultPsqlDSN,
		EncryptionKey:            DefaultServerEncryptionKey,
		AccrualAddress:           DefaultAccrualAddress,
		QueueBufSize:             DefaultQueueBufSize,
		QueueWorkers:             DefaultQueueWorkers,
		AccrualClientTImeout:     DefaultAccrualClientTImeout,
		WorkerPoolChanBuffer:     DefaultWorkerPoolChanBuffer,
		DumperStoragePath:        DefaultDumperStoragePath,
	}
}

func (c *GophermartConfig) Validate() error {
	if err := ValidateAddr(c.ListenAddr); err != nil {
		return err
	}
	if c.EncryptionKey == "" {
		return errors.New("encryption key can't be an empty string")
	}
	return nil
}

func (c *GophermartConfig) UpdateFromEnvironment() error {
	c.ListenAddr = GetEnvOrFallback("RUN_ADDRESS", c.ListenAddr)
	c.PsqlDSN = GetEnvOrFallback("DATABASE_URI", c.PsqlDSN)
	c.AccrualAddress = GetEnvOrFallback("ACCRUAL_SYSTEM_ADDRESS", c.AccrualAddress)
	c.EncryptionKey = GetEnvOrFallback("GOPHERMART_ENCRYPTION_KEY", c.EncryptionKey)
	return nil
}

func (c *GophermartConfig) UpdateFromCLIArgs(flagSet *flag.FlagSet, args []string) error {
	flagSet.StringVar(&c.PsqlDSN, "d", c.PsqlDSN, "SQL database DSN string")
	flagSet.StringVar(&c.ListenAddr, "a", c.ListenAddr, "server listen address")
	flagSet.StringVar(&c.AccrualAddress, "r", c.AccrualAddress, "accrual address")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	return nil
}

// Returns the instance of server configuration struct.
//
// Fills it's fields by using CLI arguments and environments.
// ENV or CLI argument or the default values
func NewGophermartConfig(flagSet *flag.FlagSet, args []string) (*GophermartConfig, error) {
	cfg := GetDefaultGophermartConfig()

	if err := cfg.UpdateFromCLIArgs(flagSet, args); err != nil {
		return nil, err
	}

	if err := cfg.UpdateFromEnvironment(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
