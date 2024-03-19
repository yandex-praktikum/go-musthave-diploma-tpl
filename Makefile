BUILD_DIR=./bin

SERVER_SOURCE_PATH=./cmd/gophermart/*.go
SERVER_BINARY_NAME=gophermart

ACCRUAL_SOURCE_BIN_PATH=./cmd/accrual/accrual_darwin_arm64
ACCRUAL_BINARY_NAME=accrual

.PHONY: all build run clean stop

all: build

build:
	@echo "Building the project..."
	@go build -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(SERVER_SOURCE_PATH)
	@cp $(ACCRUAL_SOURCE_BIN_PATH) $(BUILD_DIR)/$(ACCRUAL_BINARY_NAME)

run: stop build
	@echo "Running the server..."
	@touch $(BUILD_DIR)/$(SERVER_BINARY_NAME).log
	@touch $(BUILD_DIR)/$(ACCRUAL_BINARY_NAME).log
	@LOG_LEVEL="info" ENV="development" $(BUILD_DIR)/$(SERVER_BINARY_NAME) > $(BUILD_DIR)/$(SERVER_BINARY_NAME).log 2>&1 &
	@$(BUILD_DIR)/$(ACCRUAL_BINARY_NAME) > $(BUILD_DIR)/$(ACCRUAL_BINARY_NAME).log 2>&1 &
	@tail -f $(BUILD_DIR)/$(SERVER_BINARY_NAME).log $(BUILD_DIR)/$(ACCRUAL_BINARY_NAME).log

clean:
	@echo "Cleaning up..."
	@go clean
	@rm -rf $(BUILD_DIR)

stop:
	@-pkill -f $(SERVER_BINARY_NAME)
	@-pkill -f $(ACCRUAL_BINARY_NAME)

