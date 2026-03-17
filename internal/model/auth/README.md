# Auth JWT Module

Пакет `auth` предоставляет инструменты для генерации и верификации JWT-токенов с поддержкой алгоритмов **HMAC** (симметричный) и **RSA** (асимметричный). Модуль спроектирован с учетом разделения ответственности (Generator vs Verifier), что идеально подходит для микросервисной архитектуры.

## Особенности
- **Безопасность**: Принудительная проверка минимальной длины ключа для HMAC (32 байта).
- **Гибкость**: Автоматическое извлечение публичного ключа из приватного в верификаторе.
- **Тестируемость**: Внедрение зависимости времени (`now func()`) позволяет тестировать истечение срока действия токенов без ожидания.
- **Совместимость**: Использует современную библиотеку `github.com/golang-jwt/jwt/v5`.

## Примеры использования

### 1. Симметричное шифрование (HMAC)
Используется, когда один и тот же секрет известен и генератору, и верификатору.

```go
package main

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "your_project/internal/model/auth"
)

func main() {
    secret := []byte("your-32-byte-long-secret-key-here-!!!")
    ttl := 24 * time.Hour

    // Создаем генератор
    tg, _ := auth.NewTokenGenerator(jwt.SigningMethodHS256, secret, ttl)
    token, _ := tg.Generate(123)

    // Создаем верификатор
    tv, _ := auth.NewTokenVerifier(jwt.SigningMethodHS256, secret)
    userID, _ := tv.Parse(token)
}
```

### 2. Асимметричное шифрование (RSA)
Используется, когда Auth-сервис подписывает токены приватным ключом, а остальные сервисы только проверяют их публичным ключом.

```go
package main

import (
    "github.com/golang-jwt/jwt/v5"
    "your_project/internal/model/auth"
)

func main() {
    // 1. В Auth-сервисе (Генерация)
    privKey, _ := auth.LoadRSAPrivateKey("certs/private.pem")
    tg, _ := auth.NewTokenGenerator(jwt.SigningMethodRS256, privKey, time.Hour)
    token, _ := tg.Generate(456)

    // 2. В других сервисах (Верификация)
    pubKey, _ := auth.LoadRSAPublicKey("certs/public.pem")
    tv, _ := auth.NewTokenVerifier(jwt.SigningMethodRS256, pubKey)
    userID, _ := tv.Parse(token)
}
```

---

## API модуля

### Типы
- `TokenGenerator`: Отвечает за создание токенов. Содержит метод `Generate(userID int64)`.
- `TokenVerifier`: Отвечает за проверку токенов. Содержит метод `Parse(token string)`.
- `Claims`: Структура данных токена (включает `RegisteredClaims` и `UserID`).

### Функции загрузки ключей
- `LoadRSAPrivateKey(path string)`: Читает и парсит `.pem` файл приватного ключа.
- `LoadRSAPublicKey(path string)`: Читает и парсит `.pem` файл публичного ключа.

### Ошибки
- `ErrInvalidToken`: Токен некорректен, подпись не совпадает или формат нарушен.
- `ErrExpiredToken`: Срок действия токена (`exp`) истек.


## Тестирование

Для запуска тестов выполните:
```bash
go test -v ./internal/model/auth/...
```


## Генерация ключей

Для работы модуля требуются корректные ключи. Ниже приведены команды для их создания.

### 1. Генерация симметричного HMAC ключа
Для алгоритма `HS256` требуется случайная строка длиной не менее 32 байт (256 бит).

```bash
# Генерация случайной строки из 32 байт и кодирование в base64 (для конфига)
openssl rand -base64 32

# Или генерация напрямую в файл (бинарный формат)
openssl rand -out hmac.key 32
```

### 2. Генерация асимметричной пары RSA
Для алгоритма `RS256` используется пара из приватного и публичного ключей.

```bash
# 1. Генерация приватного ключа (2048 бит)
openssl genrsa -out private.pem 2048

# 2. Извлечение публичного ключа из приватного
openssl rsa -in private.pem -pubout -out public.pem

# 3. (Опционально) Просмотр структуры ключа
openssl rsa -in private.pem -text -noout
```

## Важно

> - Файл `private.pem` должен быть доступен **только** сервису авторизации. Установите права доступа: `chmod 600 private.pem`.
> - Файл `public.pem` можно безопасно распространять между всеми микросервисами, которым нужно проверять токены.
> - Никогда не фиксируйте `.pem` или `.key` файлы в Git (добавьте их в `.gitignore`).
