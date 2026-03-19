package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

// +--------------------------+
// |  Config - иммутабельный  |
// +--------------------------+
type Config struct {
	version                  string
	listenAddr               SocketAddr
	accrualSystemAddr        SocketAddr
	dbDSN                    string
	jwtSecretKey             []byte
	jwtTTL                   time.Duration
	sessTTL                  time.Duration
	maxBodySize              int64
	compressibleContentTypes map[string]struct{}
	lookupEnv                func(key string) (string, bool)
}

// NewConfig фабрика конфига, которая должна вызываться один раз
func NewConfig() Config {
	return Config{
		version:           "0.0.1",
		listenAddr:        SocketAddr{hostname: "localhost", port: "8080"},
		accrualSystemAddr: SocketAddr{hostname: "localhost", port: "8090"},
		dbDSN:             "",
		jwtSecretKey:      []byte{},
		jwtTTL:            10 * time.Minute,
		sessTTL:           30 * 24 * time.Hour,
		maxBodySize:       1024 * 1024,
		compressibleContentTypes: map[string]struct{}{
			"application/json":       {},
			"text/html":              {},
			"text/css":               {},
			"application/javascript": {},
		},

		lookupEnv: os.LookupEnv,
	}
}

type ConfigOption func(Config) (Config, error)

func (c Config) Init(cfgOpts ...ConfigOption) (Config, error) {
	var err error
	for _, opt := range cfgOpts {
		c, err = opt(c)
		if err != nil {
			return c, err
		}
	}
	return c, c.Validate()
}

func WithCmdArgs(cmdArgs *[]string) ConfigOption {
	return func(c Config) (Config, error) {
		fs := flag.NewFlagSet("config", flag.ContinueOnError)

		fs.Func("d", fmt.Sprintf("DB DSN (%s)", c.DBDSN()), func(s string) error {
			c.dbDSN = s
			return nil
		})

		fs.Func("a", fmt.Sprintf("HTTP server address (%s)", c.ListenAddr()), func(s string) error {
			sAddr, err := NewSocketAddr(s)
			if err != nil {
				return fmt.Errorf("invalid ListenAddr format: %w", err)
			}
			c.listenAddr = sAddr
			return nil
		})

		fs.Func("r", fmt.Sprintf("Accrual system server address (%s)", c.AccrualSystemAddr()), func(s string) error {
			asAddr, err := NewSocketAddr(s)
			if err != nil {
				return fmt.Errorf("invalid AccrualSystemAddr format: %w", err)
			}
			c.accrualSystemAddr = asAddr
			return nil
		})

		if cmdArgs != nil {
			if err := fs.Parse(*cmdArgs); err != nil {
				return c, fmt.Errorf("failed to parse flags: %w", err)
			}
		}

		return c, nil
	}
}

func WithEnv() ConfigOption {
	return func(c Config) (Config, error) {
		if val, ok := c.lookupEnv("DATABASE_URI"); ok {
			c.dbDSN = val
		}
		if val, ok := c.lookupEnv("JWT_SECRET_KEY"); ok {
			c.jwtSecretKey = []byte(val)
		}
		if val, ok := c.lookupEnv("RUN_ADDRESS"); ok {
			addr, err := NewSocketAddr(val)
			if err != nil {
				return c, fmt.Errorf("env RUN_ADDRESS error: %w", err)
			}
			c.listenAddr = addr
		}
		if val, ok := c.lookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
			addr, err := NewSocketAddr(val)
			if err != nil {
				return c, fmt.Errorf("env ACCRUAL_SYSTEM_ADDRESS error: %w", err)
			}
			c.accrualSystemAddr = addr
		}

		return c, nil
	}
}

func (c Config) Validate() error {
	if c.dbDSN == "" {
		return fmt.Errorf("database DSN is required")
	}
	if c.listenAddr.String() == "" {
		return fmt.Errorf("listen address is required")
	}
	if c.accrualSystemAddr.String() == "" {
		return fmt.Errorf("accrual system address is required")
	}
	if string(c.jwtSecretKey) == "" {
		return fmt.Errorf("JWT secret key is required")
	}

	return nil
}

func (c Config) JWTTTL() time.Duration {
	return c.jwtTTL
}
func (c Config) SetJWTTTL(jwtTTL time.Duration) Config {
	c.jwtTTL = jwtTTL
	return c
}

func (c Config) SessTTL() time.Duration {
	return c.sessTTL
}
func (c Config) SetSessTTL(sessTTL time.Duration) Config {
	c.sessTTL = sessTTL
	return c
}

func (c Config) JWTSecretKey() []byte {
	return c.jwtSecretKey
}
func (c Config) SetJWTSecretKey(jwtSecretKey []byte) Config {
	c.jwtSecretKey = jwtSecretKey
	return c
}

func (c Config) MaxBodySize() int64 {
	return c.maxBodySize
}
func (c Config) SetMaxBodySize(maxBodySize int64) Config {
	c.maxBodySize = maxBodySize
	return c
}

func (c Config) DBDSN() string {
	return c.dbDSN
}
func (c Config) SetDBDSN(dbDSN string) Config {
	c.dbDSN = dbDSN
	return c
}

func (c Config) CompressibleContentTypes() map[string]struct{} {
	cp := make(map[string]struct{}, len(c.compressibleContentTypes))
	for k, v := range c.compressibleContentTypes {
		cp[k] = v
	}
	return cp
}
func (c Config) SetCompressibleContentTypes(compressibleContentTypes map[string]struct{}) Config {
	cp := make(map[string]struct{}, len(compressibleContentTypes))
	for k, v := range compressibleContentTypes {
		cp[k] = v
	}
	c.compressibleContentTypes = cp
	return c
}

func (c Config) Version() string {
	return c.version
}
func (c Config) SetVersion(version string) Config {
	c.version = version
	return c
}

func (c Config) ListenAddr() string {
	return c.listenAddr.String()
}
func (c Config) SetListenAddr(listenAddr SocketAddr) Config {
	c.listenAddr = listenAddr
	return c
}

func (c Config) AccrualSystemAddr() string {
	return c.accrualSystemAddr.String()
}
func (c Config) SetAccrualSystemAddr(accrualSystemAddr SocketAddr) Config {
	c.accrualSystemAddr = accrualSystemAddr
	return c
}

// +----------------------------+
// | SocketAddr - иммутабельный |
// +----------------------------+
type SocketAddr struct {
	hostname string
	port     string
}

func NewSocketAddr(host string) (SocketAddr, error) {
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		return SocketAddr{}, fmt.Errorf("bad net address: %w", err)
	}
	return SocketAddr{hostname: hostname, port: port}, nil
}

func (s SocketAddr) String() string {
	return net.JoinHostPort(s.hostname, s.port)
}
