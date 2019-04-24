PROFILE_TEST_DB_CONN_STR ?= "postgres://postgres:docker@localhost/profiletest?sslmode=disable"

build:
	go build -o build/profile ./cmd/profile

fmt:
	gofmt -w .
	goimports -w .

test-integration: build
	PROFILE_TEST_DB_CONN_STR=$(PROFILE_TEST_DB_CONN_STR) go test -race -count=1 $(TEST_FLAGS) -tags=integration github.com/decentraland/world/internal/profile_test

vtest-integration: TEST_FLAGS=-v
vtest-integration: test-integration

tidy:
	go mod tidy

.PHONY: build
