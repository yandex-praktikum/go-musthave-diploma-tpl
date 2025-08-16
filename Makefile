.PHONY: build run test clean migrate generate-mocks

BINARY_NAME=gophermart
BUILD_DIR=bin
MIGRATIONS_DIR=migrations

# Сборка приложения
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gophermart

# Запуск приложения
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Тестирование
test: generate-mocks
	@echo "Starting PostgreSQL for tests..."
	@docker compose up -d postgres
	@sleep 3
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests completed"

# Очистка
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

# Генерация моков
generate-mocks:
	@echo "Generating mocks..."
	@if ! command -v mockery >/dev/null 2>&1; then \
		go install github.com/vektra/mockery/v2@latest; \
	fi
	@export PATH=$$PATH:$$(go env GOPATH)/bin && \
	mockery --dir internal/server --name Storage --output internal/storage/mocks --outpkg mocks --with-expecter && \
	mockery --dir internal/services --name AccrualServiceIface --output internal/services/mocks --outpkg mocks --with-expecter

# Миграции базы данных
migrate:
	@echo "Running database migrations..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URI)" up

# Создание миграции
migrate-create:
	@echo "Creating new migration..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir $(MIGRATIONS_DIR) create $(name) sql

# Установка зависимостей
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Линтинг
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

# Форматирование кода
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Проверка кода
check: fmt lint test

# Скачивание statictest
download-statictest:
	@echo "Downloading statictest..."
	./scripts/download_statictest.sh

# Скачивание автотестов
download-autotests:
	@echo "Downloading autotests..."
	./scripts/download_autotests.sh

# Запуск statictest
statictest:
	@echo "Running statictest..."
	go vet -vettool=./.tools/statictest ./...

# Запуск автотестов
autotest: build
	@echo "Running autotests..."
	@if [ ! -f .tools/gophermarttest ]; then \
		echo "gophermarttest not found, downloading..."; \
		./scripts/download_autotests.sh; \
	fi
	@if [ ! -f .tools/random ]; then \
		echo "random not found, downloading..."; \
		./scripts/download_autotests.sh; \
	fi
	@echo "Starting PostgreSQL..."
	@docker compose up -d postgres
	@sleep 5
	@echo "Running migrations..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" up
	@echo "Starting accrual server..."
	@echo $$(.tools/random unused-port) > .accrual_port
	@ACCRUAL_PORT=$$(cat .accrual_port) && echo "Accrual port: $$ACCRUAL_PORT" && RUN_ADDRESS=":$$ACCRUAL_PORT" DATABASE_URI="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" ./cmd/accrual/accrual_darwin_arm64 &
	@ACCRUAL_PID=$$!
	@sleep 2
	@echo "Starting gophermart server..."
	@ACCRUAL_PORT=$$(cat .accrual_port) && RUN_ADDRESS="localhost:8080" DATABASE_URI="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" ACCRUAL_SYSTEM_ADDRESS="http://localhost:$$ACCRUAL_PORT" ORDER_PROCESS_INTERVAL="5s" ./bin/gophermart &
	@GOPHERMART_PID=$$!
	@sleep 3
	@echo "Running gophermarttest..."
	@ACCRUAL_PORT=$$(cat .accrual_port) && .tools/gophermarttest \
		-test.v -test.run=^TestGophermart$$ \
		-gophermart-binary-path=bin/gophermart \
		-gophermart-host=localhost \
		-gophermart-port=8080 \
		-gophermart-database-uri="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" \
		-accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
		-accrual-host=localhost \
		-accrual-port=$$ACCRUAL_PORT \
		-accrual-database-uri="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	@kill $$GOPHERMART_PID 2>/dev/null || true
	@kill $$ACCRUAL_PID 2>/dev/null || true
	@rm -f .accrual_port

# Интеграционные тесты с базой данных
test-integration: generate-mocks
	@echo "Starting PostgreSQL for integration tests..."
	@docker compose up -d postgres
	@sleep 3
	@echo "Running integration tests..."
	go test -v ./internal/storage/...
	@echo "Integration tests completed"

# Полная проверка кода
check-all: fmt lint test statictest
