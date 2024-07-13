package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Резервное копирование существующих аргументов командной строки и переменных окружения.
	oldArgs := os.Args
	oldEnv := os.Environ()
	defer func() {
		os.Args = oldArgs
		os.Clearenv()
		for _, e := range oldEnv {
			pair := strings.SplitN(e, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()

	// Инициализация аргументов командной строки для разбора флагов.
	os.Args = []string{"test", "-a=:8080"}

	// Установка переменных окружения для тестирования.
	os.Setenv("ADDRESS", ":8081")

	// Создание объекта Config и загрузка его значений.
	config := NewConfig()
	config.Load()

	if *config.Port != ":8081" {
		t.Errorf("Ожидалось, что Port будет ':8081', но получено: %s", *config.Port)
	}
}

func TestLoadConfigWithNoEnvVars(t *testing.T) {
	// Резервное копирование существующих аргументов командной строки и переменных окружения.
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	// Инициализация аргументов командной строки для разбора флагов.
	os.Args = []string{"test", "-a=:8080"}

	// Создание объекта Config и загрузка его значений.
	config := NewConfig()
	config.Load()

	if *config.Port != ":8080" {
		t.Errorf("Ожидалось, что ServerPort будет ':8080', но получено: %s", *config.Port)
	}
}
