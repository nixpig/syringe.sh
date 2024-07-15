ifneq (,$(wildcard .env))
	include .env
	export
endif

SERVER_APP_PACKAGE_PATH := cmd/server/*.go
SERVER_APP_BINARY_NAME := syringeserver
CLI_APP_PACKAGE_PATH := cmd/cli/*.go
CLI_APP_BINARY_NAME := syringe

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

.PHONY: build_cli
build_cli:
	go build -o tmp/bin/${CLI_APP_BINARY_NAME} ${CLI_APP_PACKAGE_PATH}

.PHONY: build_server
build_server:
	go build -o tmp/bin/${SERVER_APP_BINARY_NAME} ${SERVER_APP_PACKAGE_PATH}

.PHONY: watch_cli
watch_cli:
		go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build_cli" \
		--build.bin "tmp/bin/${CLI_APP_BINARY_NAME}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go" \
		--misc.clean_on_exit "true"

.PHONY: watch_server
watch_server:
		go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build_server" \
		--build.bin "tmp/bin/${SERVER_APP_BINARY_NAME}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go" \
		--misc.clean_on_exit "true"

.PHONY: run_cli
run_cli: 
	go run ${CLI_APP_PACKAGE_PATH}

.PHONY: run_server
run_server: 
	go run ${SERVER_APP_PACKAGE_PATH}

.PHONY: clean
clean:
	rm -rf tmp dist *.out
	go clean

.PHONY: env
env: 
	# Echos out environment variables
	SERVER_APP_PACKAGE_PATH=${SERVER_APP_PACKAGE_PATH}
	SERVER_APP_BINARY_NAME=${SERVER_APP_BINARY_NAME}
	CLI_APP_PACKAGE_PATH=${CLI_APP_PACKAGE_PATH}
	CLI_APP_BINARY_NAME=${CLI_APP_BINARY_NAME}
	DATABASE_URL=${DATABASE_URL}
	DATABASE_TOKEN=${DATABASE_TOKEN}
	API_BASE_URL=${API_BASE_URL}
	API_TOKEN=${API_TOKEN}
	DB_ORG=${DB_ORG}
	DB_GROUP=${DB_GROUP}
	HOST_KEY_PATH=${HOST_KEY_PATH}

