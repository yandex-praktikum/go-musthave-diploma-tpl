# Gophermart — Накопительная система лояльности

Дипломный проект курса «Go-разработчик» Яндекс Практикума. Сервис предоставляет HTTP API для управления балансом лояльности пользователей, регистрации заказов и взаимодействия с системой расчета начислений.

## Badges

[![Coverage](https://img.shields.io/codecov/c/github/KalessinD/gophermart?style=flat-square)](https://codecov.io/gh/KalessinD/gophermart)

## Архитектура и Технологии

* **Go 1.26+**: Основной язык разработки.
* **PostgreSQL**: Хранилище данных.
* **Chi Router**: Маршрутизация HTTP запросов.
* **Clean Architecture**: Разделение на слои (Handlers, Services, Repositories).
* **Testcontainers**: Для интеграционного (e2e) тестирования.
* **Worker Pool**: Асинхронная обработка заказов через систему Accrual.

## Структура проекта

* `cmd/gophermart` — Точка входа сервера.
* `internal/clients` — HTTP клиенты для работы с Accrual API.
* `internal/common` — Общие конфигурации и utility функции.
* `internal/config` — Конфигурационные файлы приложения.
* `internal/handlers` — HTTP обработчики.
* `internal/gophermart` — Расширения основного приложения.
* `internal/logger` — Логгер.
* `internal/middleware` — HTTP-обёртки для обработки запросов.
* `internal/models`  — Модели данных.
* `internal/services` — Бизнес-логика.
* `internal/repositories` — Слой работы с БД (PostgreSQL) и файловым хранилищем.
* `internal/processors` — Механизм очередей и воркеры.
* `tests/e2e` — E2E тесты.

## Запуск

### 1. Конфигурация
Приложение конфигурируется через переменные окружения или флаги командной строки:

| Переменная | Флаг | Описание | Значение по умолчанию |
|---|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес и порт запуска сервиса | `:9081` |
| `DATABASE_URI` | `-d` | Строка подключения к PostgreSQL | (пусто) |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | Адрес системы расчёта начислений | (пусто) |
| `GOPHERMART_ENCRYPTION_KEY` |   | Ключ для шифрования JWT | (сгенерирован) |

### 2. Docker Compose
Для запуска инфраструктуры (БД и Accrual) используйте:

```bash
make build-docker start-docker
```

Для запуска самого сервиса (требуется Go установленный локально):
```bash
make build-gophermart start-gophermart
```

Собрать и запустить всё вместе
```bash
make build start
```

## Тестирование

### Unit-тесты
Запуск юнит-тестов
```bash
make test-go
```

### E2E тесты
Интеграционные тесты используют `testcontainers` для поднятия PostgreSQL и Accrual.

**Важно:** Для успешного прохождения E2E тестов, требующих взаимодействия с системой Accrual, необходимо наличие бинарного файла системы accrual:
- Путь: `cmd/accrual/accrual_linux_amd64`

Если бинарный файл отсутствует, тесты пададут с ошибкой `file not found in build context`.

Запуск e2e тестов:
```bash
make test-e2e
```

### Тесты от Yandex.Practicum
Запуск
```bash
make test-yp
```

Запуск конкретного теста или группы тестов
```bash
YP_CUSTOM_TEST="TestGophermart/TestUserAuth" make test-yp-custom

YP_CUSTOM_TEST="TestGophermart/TestUserOrders" make test-yp-custom

YP_CUSTOM_TEST="TestGophermart/TestEndToEnd" make test-yp-custom
```

## API Endpoints

### Публичные роуты
* `POST /api/user/register` — Регистрация пользователя.
* `POST /api/user/login` — Аутентификация пользователя.

### Приватные роуты (требуется Authorization cookie)
* `POST /api/user/orders` — Загрузка номера заказа.
* `GET /api/user/orders` — Получение списка заказов.
* `GET /api/user/balance` — Получение текущего баланса.
* `POST /api/user/balance/withdraw` — Запрос на списание средств.
* `GET /api/user/withdrawals` — Получение истории списаний.

## Makefile цели

* `make help` — Выводит список поддерживаемых целей.
* `make build` — Сборка бинарника.
* `make lint` — Запуск линтеров.
* `make lint-golangci` — Запуск линтеров golangci.
* `make lint-golangci-fix` — Запуск линтеров golangci в режиме auto-fix.
* `make lint-vet` — Запуск линтеров vet.
* `make log-gophermart` — Просмотр журналов сервера Gophermart.
* `make test` — Запуск всех тестов (unit + e2e).
* `make coverage` — ФОрмирование отчета с покрытием в формате TXT.
* `make coverage-html` — ФОрмирование отчета с покрытием в формате HTL.
* `make clean` — Удаление артефактов сборки.
* `make start` — Запуск прилоежения.
* `make stop` — Останов приложения.
* `make status` — Отображение статуса сборки контейнеров и сервера.

# Ссылки Yandex.Practicum

## Индивидуаульный проект
- [Репозиторий](https://github.com/yandex-praktikum/go-musthave-diploma-tpl).
- [Техническое задание](SPECIFICATION.md).

## Групповой проект
- [Репозиторий](https://github.com/yandex-praktikum/go-musthave-group-diploma-tpl).
- [Техническое задание](SPECIFICATION-FULL.md).