FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY --parents --from=current /cmd/gophermart/ /internal/ /migrations go.mod go.sum .golangci.json README.md ./

RUN go mod download && \
    go build -o /gophermart ./cmd/gophermart

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /gophermart .
COPY --from=builder /app/migrations ./migrations

CMD ["./gophermart"]