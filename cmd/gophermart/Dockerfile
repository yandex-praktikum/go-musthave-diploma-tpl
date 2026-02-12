# Сборка
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/gophermart/main.go

# Финальный образ
FROM golang:1.22.2-alpine

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Запуск
CMD ["./main"]