# Telemetry Metrics Module

Модуль предназначен для сбора и экспорта технических показателей приложения `gmart` в формате Prometheus.

## Структура пакета

- `metrics.go`: Общие типы и константы операций (`OpType`).
- `forpginstance/`: Метрики работоспособности и производительности PostgreSQL.
- `forauth/`: Метрики процессов авторизации и безопасности.

## Операции (OpType)

Используются для типизации меток (labels) в гистограммах и счетчиках:
- `query` / `exec`: Запросы к БД.
- `tx` / `pool`: Работа с транзакциями и пулом соединений.
- `bcrypt`: Операции хеширования и проверки паролей.

---

## Реализованные метрики

### 1. PostgreSQL (`forpginstance`)
Мониторинг состояния и производительности каждого инстанса БД.


| Метрика | Тип | Описание | Метки |
| :--- | :--- | :--- | :--- |
| `gmart_postgresql_pg_instance_ready` | Gauge | 1 если инстанс Online, 0 если Offline | `instance` |
| `gmart_postgresql_pg_instance_offline_total` | Counter | Общее количество событий ухода в Offline | `instance` |
| `gmart_postgresql_pg_instance_retries_total` | Counter | Общее количество повторных попыток | `instance`, `type` |
| `gmart_postgresql_pg_instance_duration_seconds` | Histogram | Длительность операций (бакеты: 1мс - 5с) | `instance`, `type` |

### 2. Auth (`forauth`)
Аналитика безопасности и производительности криптографических функций.


| Метрика | Тип | Описание | Метки |
| :--- | :--- | :--- | :--- |
| `gmart_auth_login_errors_total` | Counter | Ошибки входа по причинам (not_found, invalid_pwd) | `reason` |
| `gmart_auth_bcrypt_duration_seconds` | Histogram | Время работы bcrypt (бакеты: 50мс - 2с) | `type` |

---

## Использование

### Инициализация
```go
reg := prometheus.NewRegistry()

// Регистрация метрик (MustRegister внутри)
pgMetrics := forpginstance.NewForPgInstance(reg)
authMetrics := forauth.NewForAuth(reg)
```

### Запись данных
```go
// Метрики БД
pgMetrics.SetStatus("main_db:5432", true)
pgMetrics.ObserveLatency("main_db:5432", metrics.OpQuery, 0.042)

// Метрики Auth
authMetrics.IncLoginError("err_user_not_found")
authMetrics.ObserveBcrypt(0.350) // Замер времени хеширования
```
