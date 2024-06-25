#!/bin/sh

printf "Install packages gophermart...\n\n"
go mod tidy

printf "Build gophermart...\n\n"
cd cmd/gophermart \
&& rm -rf gophermart \
&& go build -buildvcs=false -o gophermart \
&& cd /app \
&& ls -la cmd/gophermart/gophermart \
&& ls -la cmd/accrual/accrual_linux_amd64 \
&& gophermarttest \
   -test.v \
   -test.run=^TestGophermart$ \
   -gophermart-binary-path=cmd/gophermart/gophermart \
   -gophermart-host=localhost \
   -gophermart-port=8080 \
   -gophermart-database-uri="postgres://postgres:postgres@postgres/praktikum?sslmode=disable" \
   -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
   -accrual-host=localhost \
   -accrual-port=8082 \
   -accrual-database-uri="postgres://postgres:postgres@postgres/praktikum?sslmode=disable"
