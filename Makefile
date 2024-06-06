ifneq (,$(wildcard .env))
	include .env
	export
endif

APP_PACKAGE_PATH := main.go
APP_BINARY_NAME := syringeserver

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
	go test -race -buildvcs -vet=off ./...

.PHONY: test
test: 
	go run gotest.tools/gotestsum@latest ./...

.PHONY: test_watch
test_watch: 
	go run gotest.tools/gotestsum@latest --watch ./...

.PHONY: coverage
coverage:
	go test -v -race -buildvcs -covermode atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: build
build:
	go build -o tmp/bin/${APP_BINARY_NAME} ${APP_PACKAGE_PATH}

.PHONY: run_watch
run_watch:
		go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" \
		--build.bin "tmp/bin/${APP_BINARY_NAME}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go" \
		--misc.clean_on_exit "true"


.PHONY: run
run: 
	go run ${APP_PACKAGE_PATH}

.PHONY: clean
clean:
	rm -rf tmp
	rm *.out

.PHONY: env
env: 
	# Echos out environment variables
	APP_PACKAGE_PATH=${APP_PACKAGE_PATH}
	APP_BINARY_NAME=${APP_BINARY_NAME}
	DATABASE_URL=${TURSO_DATABASE_URL}
	DATABASE_TOKEN=${TURSO_AUTH_TOKEN}
