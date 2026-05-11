ifneq (,$(wildcard .env))
include .env
export YP_PSQL_DSN
export YP_CUSTOM_TEST
export DOCKER_COMPOSE/check
endif

SHELL := /bin/bash
PROJECT_DIR ?= $(CURDIR)
TMPDIR ?= /tmp

# Конфигурация
YP_AUTOTESTS_GIT_URL ?= "ssh://git@github.com/Yandex-Practicum/go-autotests"
YP_AUTOTESTS_PATH ?= $(TMPDIR)/go-autotests
YP_GOPHERMART_TEST ?= $(HOME)/bin/gophermarttest
YP_CUSTOM_TEST ?= ""
YP_PSQL_DSN ?= ""

GO_PACKAGES := $(shell go list ./... | grep -vE '/mocks|/e2e')
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

ifeq ($(OS),darwin)
    ARCH := $(shell uname -m)
else
    ARCH := amd64
endif

# Инструменты
GOLANGCI_LINT ?= golangci-lint
GO ?= go
GIT := git
GREP := grep
RM := rm -f
RMDIR := rm -rf
CD := cd
TAIL := tail
KILL := kill
SLEEP := sleep
DOCKER := docker
DOCKER_COMPOSE ?= $(DOCKER) compose
ECHO := echo -e

NOECHO := @

TAIL_LAST_N_LINES ?= 10

# Логи и PID
ACCRUAL_LOG_FILE := $(TMPDIR)/practicum-accrual.log
ACCRUAL_PID_FILE := $(TMPDIR)/practicum-accrual.pid
ACCRUAL_CMD := $(PROJECT_DIR)/cmd/accrual
ACCRUAL_BIN := $(ACCRUAL_CMD)/accrual_$(OS)_$(ARCH)
ACCRUAL_HOST ?= localhost
ACCRUAL_PORT ?= 9081
ACCRUAL_ADDRESS := http://$(ACCRUAL_HOST):$(ACCRUAL_PORT)

GOPHERMART_LOG_FILE := $(TMPDIR)/practicum-gophermart.log
GOPHERMART_PID_FILE := $(TMPDIR)/practicum-gophermart.pid
GOPHERMART_CMD := $(PROJECT_DIR)/cmd/gophermart
GOPHERMART_BIN := $(GOPHERMART_CMD)/gophermart
GOPHERMART_HOST ?= localhost
GOPHERMART_PORT ?= 9082

GO_COVERAGE_REPORT := $(TMPDIR)/practicum-gophermart-coverage.out

# Красивый вывод
print_title = $(ECHO) "\033[1;33m$1\033[0m"

.PHONY: all help clean \
	build build-gophermart build-docker \
	test test-go test-e2e test-yp-iterations test-yp-custom \
	lint lint-vet lint-golangci lint-golangci-fix \
	coverage coverage-html \
	clone-yp-autotest \
	check-binaries \
#	build-accrual start-accrual stop-accrual status-accrual log-accrual
	start-gophermart stop-gophermart log-gophermart status-gophermart \
	start-docker stop-docker \
	start stop restart status

.DEFAULT_GOAL := all

.ONESHELL:

all: stop clean build lint test-go  # Builds gophermart and accrual binaries, runs tests

help: # Shows help message
	$(NOECHO) $(GREP) -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | \
	sort | \
	while read -r l; do \
		printf "\033[1;33m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; \
	done

clone-yp-autotest: # Clones Yandex.Practicum auto-test from git repository
	$(NOECHO) $(call print_title,"Cloning Yandex.Practicum auto-test from git repository")
	$(NOECHO) if [ ! -e "$(YP_AUTOTESTS_PATH)" ]; then \
		$(ECHO) "Cloning tests"; \
		$(GIT) clone $(YP_AUTOTESTS_GIT_URL) $(YP_AUTOTESTS_PATH); \
	fi

build: build-gophermart build-docker # Builds gophermart and accrual binaries

build-docker: # Builds docker compose
	$(NOECHO) $(DOCKER_COMPOSE) -f docker-compose.yml build

#build-accrual: # Builds accrual's binary
#	$(NOECHO) $(call print_title,"Building accrual binary")
#	$(NOECHO) $(GO) build -o $(ACCRUAL_BIN) $(ACCRUAL_CMD)

build-gophermart: # Builds gophermart's binary
	$(NOECHO) $(call print_title,"Building gophermart binary")
	$(NOECHO) $(GO) build -o $(GOPHERMART_BIN) $(GOPHERMART_CMD)

clean: # Removes binaries and logs
	$(NOECHO) $(call print_title,"Removing built binaries and checkouted tests")
	$(NOECHO) $(RM) $(GOPHERMART_BIN) \
		$(GOPHERMART_LOG_FILE) \
		$(GO_COVERAGE_REPORT)
#		$(ACCRUAL_BIN) \
#		$(ACCRUAL_LOG_FILE) \


lint: lint-vet lint-golangci # Runs linters from govet and golangci-lint respectively

lint-vet: # Runs go vet with structtag check
	$(NOECHO) $(call print_title,"Running go vet with structtag check")
	$(NOECHO) $(GO) vet -structtag ./...

lint-golangci: # Runs linters from golangci-lint
	$(NOECHO) $(call print_title,"Running golangci linters")
	$(NOECHO) $(GOLANGCI_LINT) run

lint-golangci-fix: # Runs golangci-lint with auto-fix
	$(NOECHO) $(GOLANGCI_LINT) run --fix

check-binaries: # Checks the existance of required binaries
	$(NOECHO) $(call print_title,"Looking up for binaries")
	$(NOECHO) if [ ! -f $(ACCRUAL_BIN) -o ! -f $(GOPHERMART_BIN) ]; then \
		$(ECHO) "accrual and gophermart binaries were not found"; \
		exit 1; \
	fi

test: test-go test-e2e test-yp # Runs tests

test-go: # Runs golang tests
	$(NOECHO) $(call print_title,"Running tests: golang")
	$(NOECHO) $(GO) clean -testcache
	$(NOECHO) $(GO) test -buildvcs=false -v -race -cover ./...

test-e2e: # Runs end2end tests
	$(NOECHO) $(call print_title,"Running e2e tests")
	$(NOECHO) $(GO) test -buildvcs=false -v -tags=e2e ./tests/...

test-yp: check-binaries stop start
	$(NOECHO) $(call print_title,"Running Yandex.Practicum tests")
	$(NOECHO) $(YP_GOPHERMART_TEST) -test.v \
		-accrual-binary-path $(ACCRUAL_BIN) \
		-gophermart-binary-path $(GOPHERMART_BIN) \
		-gophermart-database-uri $(YP_PSQL_DSN) \
		-accrual-database-uri $(YP_PSQL_DSN) \
		-gophermart-host $(GOPHERMART_HOST) \
		-gophermart-port $(GOPHERMART_PORT) \
		-accrual-host $(ACCRUAL_HOST) \
		-accrual-port $(ACCRUAL_PORT)

test-yp-custom: clone-yp-autotest check-binaries stop start # Runs Yandex.Practicum test cases
	$(NOECHO) $(call print_title,"Running Yandex.Practicum tests for custom iteration")
	$(NOECHO) if [ -z "${YP_CUSTOM_TEST}" ]; then \
		$(ECHO) "Please set YP_CUSTOM_TEST variable to the name of the test you want to run (e.g. 'TestIteration8/TestGetGzipHandlers/get_info_page')"; \
		exit 1; \
	fi
	$(NOECHO) $(CD) $(YP_AUTOTESTS_PATH); \
		$(GO) clean -testcache; \
		$(GO) test -v -run ${YP_CUSTOM_TEST} \
		./cmd/gophermarttest/ \
		-accrual-binary-path $(ACCRUAL_BIN) \
		-gophermart-binary-path $(GOPHERMART_BIN) \
		-gophermart-database-uri $(YP_PSQL_DSN) \
		-accrual-database-uri $(YP_PSQL_DSN) \
		-gophermart-host $(GOPHERMART_HOST) \
		-gophermart-port $(GOPHERMART_PORT) \
		-accrual-host $(ACCRUAL_HOST) \
		-accrual-port $(ACCRUAL_PORT)

coverage: # Runs tests and shows total coverage
	$(NOECHO) $(call print_title,"Running tests with coverage")
	$(NOECHO) $(GO) test -buildvcs=false -v -race -coverprofile=$(GO_COVERAGE_REPORT) $(GO_PACKAGES)
	$(NOECHO) $(GO) tool cover -func=$(GO_COVERAGE_REPORT)

coverage-html: # Generates HTML coverage report and opens it
	$(NOECHO) $(call print_title,"Generating HTML coverage report")
	$(NOECHO) $(GO) test -v -race -coverprofile=$(GO_COVERAGE_REPORT) $(GO_PACKAGES)
	# $(NOECHO) $(GO) test -v -race -coverprofile=$(GO_COVERAGE_REPORT) ./...
	$(NOECHO) $(GO) tool cover -html=$(GO_COVERAGE_REPORT)

start: # Starts the server and the accrual to communicate each other
	$(NOECHO) $(MAKE) start-docker
	$(NOECHO) $(SLEEP) 3 # giving some time to start
	$(NOECHO) $(MAKE) start-gophermart
#	$(NOECHO) $(SLEEP) 3 # giving some time to start
#	$(NOECHO) $(MAKE) start-accrual
	
start-docker: # Starts the Docker container with PostgreSQL DB
	$(NOECHO) $(call print_title,"Starting up the docker compose containers")
	$(NOECHO) $(DOCKER_COMPOSE) up -d

#start-accrual: # Starts the accrual
#	$(NOECHO) $(call print_title,"Starts the accrual")
#	$(NOECHO) if [ -f $(ACCRUAL_PID_FILE) ] && $(KILL) -0 $$(cat $(ACCRUAL_PID_FILE)) 2>/dev/null; then \
#		$(ECHO) "accrual is already running (PID: $$(cat $(ACCRUAL_PID_FILE)))"; \
#		exit 1; \
#	fi
#	$(NOECHO) $(ACCRUAL_BIN) \
#		-d $(YP_PSQL_DSN) \
#		-a $(ACCRUAL_HOST):$(ACCRUAL_PORT) \
#		> $(ACCRUAL_LOG_FILE) 2>&1 & $(ECHO) $$! > $(ACCRUAL_PID_FILE)

start-gophermart: # Starts the gophermart
	$(NOECHO) $(call print_title,"Starts the gophermart")
	$(NOECHO) if [ -f $(GOPHERMART_PID_FILE) ] && $(KILL) -0 $$(cat $(GOPHERMART_PID_FILE)) 2>/dev/null; then \
		$(ECHO) "Gophermart is already running (PID: $$(cat $(GOPHERMART_PID_FILE)))"; \
		exit 1; \
	fi
	$(NOECHO) $(GOPHERMART_BIN) \
		-d $(YP_PSQL_DSN) \
		-a :$(GOPHERMART_PORT) \
		-r $(ACCRUAL_ADDRESS) \
		> $(GOPHERMART_LOG_FILE) 2>&1 & $(ECHO) $$! > $(GOPHERMART_PID_FILE)

stop: stop-gophermart stop-docker # Stops the gophermart, the accrual and docker containers

stop-docker: # Stops the Docker container with PostgreSQL DB
	$(NOECHO) $(call print_title,"Stopping the docker compose containers")
	$(NOECHO) $(DOCKER_COMPOSE) down

#stop-accrual: # Stops the accrual
#	$(NOECHO) $(call print_title,"Stops the accrual")
#	$(NOECHO) if [ -f $(ACCRUAL_PID_FILE) ]; then \
#		PID=$$(cat $(ACCRUAL_PID_FILE)); \
#		$(KILL) $$PID; \
#		$(RM) $(ACCRUAL_PID_FILE); \
#		$(ECHO) "Stopped accrual (PID $$PID)"; \
#	else \
#		$(ECHO) "accrual is not running"; \
#	fi

stop-gophermart: # Stops the gophermart
	$(NOECHO) $(call print_title,"Stops the gophermart")
	$(NOECHO) if [ -f $(GOPHERMART_PID_FILE) ]; then \
		PID=$$(cat $(GOPHERMART_PID_FILE)); \
		$(KILL) $$PID; \
		$(RM) $(GOPHERMART_PID_FILE); \
		$(ECHO) "Stopped gophermart (PID $$PID)"; \
	else \
		echo "Gophermart is not running"; \
	fi

restart: stop start # Restarts services

#log-accrual: # Shows log from accrual
#	$(TAIL) -n $(TAIL_LAST_N_LINES) $(ACCRUAL_LOG_FILE)

log-gophermart: # Shows log from gophermart
	$(TAIL) -n $(TAIL_LAST_N_LINES) $(GOPHERMART_LOG_FILE)

status: status-gophermart # Returns the status of gophermart and accrual
	$(NOECHO) $(DOCKER_COMPOSE) ps -a

#status-accrual: # Returns the status of the accrual
#	$(NOECHO) if [ -f $(ACCRUAL_PID_FILE) ] && $(KILL) -0 $$(cat $(ACCRUAL_PID_FILE)) 2>/dev/null; then \
#		$(ECHO) "accrual is running (PID $$(cat $(ACCRUAL_PID_FILE)))"; \
#	else \
#		$(ECHO) "accrual is stopped"; \
#	fi

status-gophermart: # Returns the status of the gophermart
	$(NOECHO) if [ -f $(GOPHERMART_PID_FILE) ] && $(KILL) -0 $$(cat $(GOPHERMART_PID_FILE)) 2>/dev/null; then \
		$(ECHO) "Gophermart is running (PID $$(cat $(GOPHERMART_PID_FILE)))"; \
	else \
		$(ECHO) "Gophermart is stopped"; \
	fi
