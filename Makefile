PROTOC ?= protoc
PROFILE_TEST_DB_CONN_STR ?= "postgres://postgres:password@localhost/profiletest?sslmode=disable"

build:
	go build -o build/profile ./cmd/profile
	go build -o build/coordinator ./cmd/comms/coordinator
	go build -o build/server ./cmd/comms/server
	go build -o build/identity ./cmd/identity
	go build -o build/worlddef ./cmd/worlddef

buildall: build
	go build -o build/cli_bot ./cmd/cli/bot
	go build -o build/cli_keygen ./cmd/cli/keygen
	go build -o build/cli_profile ./cmd/cli/profile

fmt:
	gofmt -w .
	goimports -w .

test:
	go test -race $(TEST_FLAGS) ./... -count=1
	PROFILE_TEST_DB_CONN_STR=$(PROFILE_TEST_DB_CONN_STR) go test -race -count=1 $(TEST_FLAGS) -tags=integration github.com/decentraland/world/internal/profile_test

profileci:
	PROFILE_TEST_DB_CONN_STR=$(PROFILE_TEST_DB_CONN_STR) go test -race -count=1 $(TEST_FLAGS) -tags=integration github.com/decentraland/world/internal/profile_test

identityci:
	go test -v ./... -count=1

tidy:
	go mod tidy

compile-protocol:
	cd pkg/protocol; ${PROTOC} --js_out=import_style=commonjs,binary:. --ts_out=. --go_out=. ./comms.proto

.PHONY: build
