# Переменные проекта
WORKDIR=.
BINARY_NAME=${WORKDIR}/cmd/gophermart/gophermart
MAIN_PATH=${WORKDIR}/cmd/gophermart/main.go # Путь может отличаться, проверь его
COVER_FILE=coverage.out
DOCKER_IMG=shortener:latest

.PHONY: all build run test cover clean lint help

# Команда по умолчанию
all: test build

build:
	@echo "Building binary..."
	go generate ./...
	CGO_ENABLED=0 GOOS=linux go build -o $(BINARY_NAME) $(MAIN_PATH)

## Docker-build: Сборка образа на базе пустого scratch
docker-build: build
	@echo "Building scratch Docker image..."
	docker build -f Dockerfile -t $(DOCKER_IMG) .

docker-run:
	@echo "Starting scratch container on port 8090..."
	docker run --rm --network host --env-file .env $(DOCKER_IMG)

## Run: Быстрый запуск приложения
run:
	@echo "Starting application..."
	go run $(MAIN_PATH)

## Test: Запуск всех тестов с проверкой на Race Condition
test:
	@echo "Running tests with race detector..."
	go test -v -race ./...

## Cover: Проверка покрытия тестами и генерация отчета
cover:
	@echo "Checking test coverage..."
	go test -coverprofile=$(COVER_FILE) ./...
	go tool cover -func=$(COVER_FILE)
	@echo "Generating HTML report..."
	go tool cover -html=$(COVER_FILE) -o coverage.html
	@echo "Report saved to coverage.html"

## Bench: Запуск бенчмарков с анализом аллокаций
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

## Escape: Анализ "побега" в кучу для всего проекта (исключая тесты)
escape:
	@echo "Analyzing escape analysis for all modules..."
	go build -gcflags="-m" ./... 2>&1 | grep -v "_test.go" | grep -E "escapes to heap|moved to heap"


## Lint: Проверка качества кода (требует golangci-lint)
lint:
	@echo "Running linter..."
	./bin/golangci-lint run

## Clean: Удаление временных файлов и бинарников
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f $(COVER_FILE)
	rm -f coverage.html
	rm -f ./storage.json
	@echo "Resetting test cache..."
	go clean -testcache

## Doc: Запуск локального сервера документации (godoc)
doc:
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Run 'go install golang.org/x/tools/cmd/godoc@latest' if not found"
	godoc -http=:6060

## Help: Список доступных команд
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

iter1:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration1$$ -binary-path=${BINARY_NAME}

iter2:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration2$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal

iter3:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration3$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal

iter4:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration4$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080

iter5:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration5$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080

iter6:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration6$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080

iter7:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration7$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080

iter8:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration8$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080

iter9:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration9$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json

iter10:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration10$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json -database-dsn='postgres://murl:murl_pswd@localhost:5530/murl_00?pool_max_conns=10'

iter11:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration11$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json -database-dsn='postgres://murl:murl_pswd@localhost:5530/murl_00'

iter12:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration12$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json -database-dsn='postgres://murl:murl_pswd@localhost:5530/murl_00'

iter13:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration13$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json -database-dsn='postgres://murl:murl_pswd@localhost:5530/murl_00'

iter14:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration14$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json -database-dsn='postgres://murl:murl_pswd@localhost:5530/murl_00'

iter15:
	${WORKDIR}/cmd/tests/shortenertest_v2 -test.v -test.run=^TestIteration15$$ -binary-path=${BINARY_NAME} -source-path=${WORKDIR}/internal -server-port=8080 -file-storage-path=./storage.json -database-dsn='postgres://murl:murl_pswd@localhost:5530/murl_00'
