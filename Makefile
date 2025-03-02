ifneq (,$(wildcard .env))
	include .env
	export
endif

BINARY_NAME := syringe
PACKAGE_PATH := cmd/cli/*.go
CGO_ENABLED=1 # required for mattn/go-sqlite3

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

.PHONY: build
build:
	go build -o tmp/bin/${BINARY_NAME}

.PHONY: run
run: 
	go run ./...

.PHONY: clean
clean:
	rm -rf tmp *.out
	go clean

.PHONY: env
env: 
	# Environment variables
	PACKAGE_PATH=${PACKAGE_PATH}
	BINARY_NAME=${BINARY_NAME}
