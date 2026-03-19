
# Переменные проекта
PRJ           := gmart
CON_NAME      := $(PRJ)_db
IMAGE_NAME    := $(PRJ)_postgres

# Настройки БД
POSTGRES_USER     := gmart
POSTGRES_PASSWORD := gmart_pswd
POSTGRES_DB       := gmart
PG_PORT           := 5550

.PHONY: build up down info logs psql

# Сборка образа
build:
	docker build -t $(IMAGE_NAME) .

# Запуск: хостовая сеть, данные в RAM или временном слое (удаляются при down)
up: build
	@status=$$(docker inspect -f '{{.State.Status}}' $(CON_NAME) 2>/dev/null || echo "none"); \
	if [ "$$status" = "running" ]; then \
		echo "===> Container $(CON_NAME) is already running."; \
	elif [ "$$status" = "exited" ] || [ "$$status" = "paused" ]; then \
		echo "===> Starting existing container $(CON_NAME)..."; \
		docker start $(CON_NAME); \
	else \
		echo "===> Creating new container $(CON_NAME)..."; \
		docker run --detach \
			--name $(CON_NAME) \
			--network host \
			-e POSTGRES_USER=$(POSTGRES_USER) \
			-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
			-e POSTGRES_DB=$(POSTGRES_DB) \
			$(IMAGE_NAME) postgres -p $(PG_PORT); \
	fi

# Остановка и полное удаление контейнера (данные теряются)
down: stop
	@docker rm $(CON_NAME) 2>/dev/null || true
	@echo "===> Container removed (data lost)."

start:
	@echo "===> Starting existing container $(CON_NAME)..."
	docker start $(CON_NAME)

stop:
	@docker stop $(CON_NAME) 2>/dev/null || true

# Информация о системе
info:
	@echo "--- Containers ---"
	@docker ps -a -f name=$(CON_NAME)
	@echo "\n--- Network (Host mode active) ---"
	@netstat -tuln | grep $(PG_PORT) || echo "Port $(PG_PORT) is not active"

# Просмотр логов
logs:
	docker logs -f $(CON_NAME)

# Подключение к БД внутри контейнера
psql:
	docker exec -it $(CON_NAME) psql -p 5550 -U $(POSTGRES_USER) -d $(POSTGRES_DB)

