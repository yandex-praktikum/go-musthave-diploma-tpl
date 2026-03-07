# Auth Repository Module

Пакет `auth` предоставляет реализацию репозитория для управления пользователями и сессиями (Refresh Tokens) в базе данных PostgreSQL с использованием драйвера `pgx/v5`.

## Функционал

*   **SignUp**: Регистрация пользователей с хешированием пароля через `bcrypt`.
*   **SignIn**: Аутентификация и создание уникальных сессий (`uuid`).
*   **RefreshSession**: Продление жизни сессии с проверкой срока годности.
*   **SignOut**: Инвалидация (удаление) сессий.
*   **Тестируемость**: Поддержка внедрения зависимости времени (`now func()`) для точных unit-тестов.

## Использование

### Инициализация

```go
repo := auth.NewAuthRepo(pgInstance, 30 * 24 * time.Hour)
```

### Примеры методов

```go
ctx := context.Background()

// Регистрация
userID, err := repo.SignUp(ctx, "admin", "secret_password")

// Вход в систему
sessionID, userID, err := repo.SignIn(ctx, "admin", "secret_password", "Mozilla/5.0", "127.0.0.1")

// Обновление сессии
userID, err := repo.RefreshSession(ctx, sessionID, "Mozilla/5.0", "127.0.0.1")
```

## Ошибки
Пакет возвращает специфичные ошибки для обработки в сервисном слое:

`ErrUserNotFound / ErrUserAlreadyExists`

`ErrInvalidPassword`

`ErrSessionExpired / ErrSessionNotFound`