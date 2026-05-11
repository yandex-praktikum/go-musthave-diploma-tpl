FROM debian:bookworm-slim

WORKDIR /app

COPY /cmd/accrual/accrual_linux_amd64 ./
# COPY --from=bin accrual_linux_amd64 ./

CMD ["./accrual_linux_amd64"]
