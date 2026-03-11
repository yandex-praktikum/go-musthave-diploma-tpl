# Запуск приложения локально через docker-compose в фоновом режиме
local:
	docker-compose -f docker-compose.yml up -d

# Генерация моков (требует mockgen в PATH).
generate:
	go generate ./internal/repository/...

# Запуск всех тестов (перед этим поднять БД: make local).
test:
	go test ./... -v
