PROTOC ?= protoc
VERSION := $(shell git rev-list -1 HEAD)
BUILD_FLAGS = -ldflags '-X github.com/decentraland/world/internal/commons/version.version=$(VERSION)'

build:
	go build $(BUILD_FLAGS) -o build/coordinator ./cmd/comms/coordinator
	go build $(BUILD_FLAGS) -o build/server ./cmd/comms/server

buildperftest:
	go build -o build/densetest ./cmd/comms/densetest
	go build -o build/sparsetest ./cmd/comms/sparsetest
	go build -o build/realistictest ./cmd/comms/realistictest

buildcli:
	go build -o build/cli_bot ./cmd/cli/bot
	go build -o build/cli_profile ./cmd/cli/profile

buildall: build buildperftest buildcli

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
