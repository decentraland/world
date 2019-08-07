PROTOC ?= protoc
VERSION := $(shell git rev-list -1 HEAD)
BUILD_FLAGS = -ldflags '-X github.com/decentraland/world/internal/commons/version.version=$(VERSION)'

build:
	go build $(BUILD_FLAGS) -o build/coordinator ./cmd/comms/coordinator
	go build $(BUILD_FLAGS) -o build/server ./cmd/comms/server
	go build $(BUILD_FLAGS) -o build/test ./cmd/comms/test
	go build -o build/cli_bot ./cmd/cli/bot
	go build -o build/cli_profile ./cmd/cli/profile

fmt:
	gofmt -w .
	goimports -w .

test:
	go test -race $(TEST_FLAGS) ./... -count=1

tidy:
	go mod tidy

compile-protocol:
	cd pkg/protocol; ${PROTOC} --js_out=import_style=commonjs,binary:. --ts_out=. --go_out=. ./comms.proto

.PHONY: build
