build:
	go build -o cmd/agent -v cmd/agent/main.go
	mv cmd/agent/main cmd/agent/agent

start:
	cmd/gophermart


.PHONY: run
run:
	air

test:
	go test -count=1 ./...

test-coverage:
	go test -coverprofile cover.out ./...

test-coverage-html: test-coverage
	go tool cover -html=cover.out

test-coverage-p: test-coverage
	go tool cover -func cover.out | fgrep total | awk '{print $3}'

test-autotests: build
	autotests/metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$$ -binary-path=cmd/server/server

db-migration-create:
	migrate create -ext sql -dir internal/config/db/migrations -seq create_users_table

db-migration-up:
	export POSTGRESQL_URL="postgres://postgres:postgres@localhost:5432/yandex_metrics?sslmode=disable"
	migrate -database ${POSTGRESQL_URL} -path internal/config/db/migrations up

db-migration-down:
	export POSTGRESQL_URL="postgres://postgres:postgres@localhost:5432/yandex_metrics?sslmode=disable"
	migrate -database ${POSTGRESQL_URL} -path internal/config/db/migrations down

.DEFAULT_GOAL := run
