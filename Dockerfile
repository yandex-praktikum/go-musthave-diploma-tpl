# Сборка приложения
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go build -o gophermart ./cmd/gophermart

# Финальный образ
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/gophermart .
COPY migrations ./migrations

ENTRYPOINT ["./gophermart"]
