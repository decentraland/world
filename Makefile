PROTOC ?= protoc
PROFILE_TEST_DB_CONN_STR ?= "postgres://postgres:docker@localhost/profiletest?sslmode=disable"

build:
	go build -o build/profile ./cmd/profile
	go build -o build/bots ./cmd/comms/bots
	go build -o build/coordinator ./cmd/comms/coordinator
	go build -o build/server ./cmd/comms/server
	go build -o build/identity ./cmd/identity
	go build -o build/keygen ./cmd/keygen

fmt:
	gofmt -w .
	goimports -w .

test:
	go test -v ./... -count=1
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
