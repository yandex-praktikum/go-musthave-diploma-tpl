package config

import (
	"flag"
	"os"
	"testing"
)

const testSecretKey = "test-secret"

// setupConfigTest готовит окружение для теста конфигурации:
// очищает env, сбрасывает и парсит флаги, устанавливает SECRET_KEY.
func setupConfigTest(t *testing.T, args ...string) {
	t.Helper()
	clearEnv(t)
	resetFlags(t, args...)
	t.Setenv("SECRET_KEY", testSecretKey)
}

// resetFlags сбрасывает глобальный набор флагов и аргументы командной строки,
// чтобы каждый тест вызывал Load() с чистым состоянием.
func resetFlags(t *testing.T, args ...string) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append([]string{"gophermart"}, args...)
}

// clearEnv удаляет переменные окружения, влияющие на конфигурацию,
// и восстанавливает их после завершения теста.
func clearEnv(t *testing.T) {
	t.Helper()
	vars := []string{"RUN_ADDRESS", "DATABASE_URI", "ACCRUAL_SYSTEM_ADDRESS", "SECRET_KEY"}
	saved := make(map[string]string)
	for _, v := range vars {
		saved[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			os.Setenv(k, v)
		}
	})
}

func TestLoadDefaults(t *testing.T) {
	setupConfigTest(t)

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.RunAddress != "localhost:8080" {
		t.Errorf("RunAddress = %q, want %q", cfg.RunAddress, "localhost:8080")
	}
	if cfg.DatabaseURI != "" {
		t.Errorf("DatabaseURI = %q, want empty", cfg.DatabaseURI)
	}
	if cfg.AccrualSystemAddress != "" {
		t.Errorf("AccrualSystemAddress = %q, want empty", cfg.AccrualSystemAddress)
	}
	if cfg.SecretKey != testSecretKey {
		t.Errorf("SecretKey = %q, want %q", cfg.SecretKey, testSecretKey)
	}
}

func TestLoadFromFlags(t *testing.T) {
	setupConfigTest(t, "-a", "host:9000", "-d", "postgres://localhost/db", "-r", "http://accrual:8000")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.RunAddress != "host:9000" {
		t.Errorf("RunAddress = %q, want %q", cfg.RunAddress, "host:9000")
	}
	if cfg.AccrualSystemAddress != "http://accrual:8000" {
		t.Errorf("AccrualSystemAddress = %q, want %q", cfg.AccrualSystemAddress, "http://accrual:8000")
	}
	// localhost заменяется на 127.0.0.1
	if cfg.DatabaseURI != "postgres://127.0.0.1/db" {
		t.Errorf("DatabaseURI = %q, want %q", cfg.DatabaseURI, "postgres://127.0.0.1/db")
	}
}

func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)
	resetFlags(t)

	t.Setenv("RUN_ADDRESS", "env:8081")
	t.Setenv("DATABASE_URI", "postgres://dbhost/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual-env:8000")
	t.Setenv("SECRET_KEY", "env-secret")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.RunAddress != "env:8081" {
		t.Errorf("RunAddress = %q, want %q", cfg.RunAddress, "env:8081")
	}
	if cfg.DatabaseURI != "postgres://dbhost/db" {
		t.Errorf("DatabaseURI = %q, want %q", cfg.DatabaseURI, "postgres://dbhost/db")
	}
	if cfg.AccrualSystemAddress != "http://accrual-env:8000" {
		t.Errorf("AccrualSystemAddress = %q, want %q", cfg.AccrualSystemAddress, "http://accrual-env:8000")
	}
	if cfg.SecretKey != "env-secret" {
		t.Errorf("SecretKey = %q, want %q", cfg.SecretKey, "env-secret")
	}
}

func TestLoadFlagsOverrideEnv(t *testing.T) {
	setupConfigTest(t, "-a", "host:9000")
	t.Setenv("RUN_ADDRESS", "env:8081")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Флаг должен перекрывать env
	if cfg.RunAddress != "host:9000" {
		t.Errorf("RunAddress = %q, want %q (flag should override env)", cfg.RunAddress, "host:9000")
	}
}

func TestLoadDockerHostNotReplaced(t *testing.T) {
	setupConfigTest(t, "-d", "postgres://host.docker.internal/db")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.DatabaseURI != "postgres://host.docker.internal/db" {
		t.Errorf("DatabaseURI = %q, want unchanged %q", cfg.DatabaseURI, "postgres://host.docker.internal/db")
	}
}
