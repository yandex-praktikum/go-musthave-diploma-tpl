# FROM golang:1.26-alpine
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=bin accrual_linux_amd64 ./

CMD ["./accrual_linux_amd64"]
