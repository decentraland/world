build:
	go build -o build/profile ./cmd/profile

fmt:
	gofmt -w .
	goimports -w .

.PHONY: build
