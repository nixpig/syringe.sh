ifneq (,$(wildcard .env))
	include .env
	export
endif

SERVER_BINARY_NAME := server
CLI_BINARY_NAME := syringe
SERVER_PACKAGE_PATH := cmd/server/*.go
CLI_PACKAGE_PATH := cmd/cli/*.go
BUILD_OUTPUT_DIR := tmp/bin
CGO_ENABLED := 1 # required for mattn/go-sqlite3

.PHONY: tidy
tidy: 
	go fmt ./...
	go mod tidy -v

.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: test
test: 
	go run gotest.tools/gotestsum@latest ./...

.PHONY: watch_test
watch_test: 
	go run gotest.tools/gotestsum@latest --watch ./...

.PHONY: coverage
coverage:
	go test -v -race -buildvcs -covermode atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: coveralls
coveralls:
	go run github.com/mattn/goveralls@latest -coverprofile=coverage.out -service=github

.PHONY: server
server:
	go build -o ${BUILD_OUTPUT_DIR}/${SERVER_BINARY_NAME} ${SERVER_PACKAGE_PATH}

.PHONY: cli
cli:
	go build -o ${BUILD_OUTPUT_DIR}/${CLI_BINARY_NAME} ${CLI_PACKAGE_PATH}

.PHONY: watch
watch:
		go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make server" \
		--build.bin "tmp/bin/${SERVER_BINARY_NAME}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go" \
		--misc.clean_on_exit "true"

.PHONY: clean
clean:
	rm -rf tmp *.out
	go clean

.PHONY: env
env: 
	# Environment variables
	SERVER_BINARY_NAME=${SERVER_BINARY_NAME}
	CLI_BINARY_NAME=${CLI_BINARY_NAME}
	SERVER_PACKAGE_PATH=${SERVER_PACKAGE_PATH}
	CLI_PACKAGE_PATH=${CLI_PACKAGE_PATH}
	BUILD_OUTPUT_DIR=${BUILD_OUTPUT_DIR}
	CGO_ENABLED=${CGO_ENABLED}
	SYRINGE_PORT=${SYRINGE_PORT}
	SYRINGE_PORT=${SYRINGE_PORT}
	SYRINGE_KEY=${SYRINGE_KEY}
	SYRINGE_DB_SYSTEM_DIR=${SYRINGE_DB_SYSTEM_DIR}
	SYRINGE_DB_SYSTEM_USER=${SYRINGE_DB_SYSTEM_USER}
	SYRINGE_DB_SYSTEM_PASSWORD=${SYRINGE_DB_SYSTEM_PASSWORD}
	SYRINGE_DB_TENANT_DIR=${SYRINGE_DB_TENANT_DIR}

