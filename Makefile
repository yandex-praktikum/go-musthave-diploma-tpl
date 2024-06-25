include .env

check:
	@echo ${DATABASE_URI}

migrate-create:
	@(printf "Enter migrate name: "; read arg; migrate create -ext sql -dir db/migrations -seq $$arg);

migrate-up:
	migrate -database ${DATABASE_URI} -path ./db/migrations up

migrate-down:
	migrate -database ${DATABASE_URI} -path ./db/migrations down 1

migrate-down-all:
	migrate -database ${DATABASE_URI} -path ./db/migrations down -all
