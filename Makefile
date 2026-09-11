.PHONY: build test vet fmt install verify

BIN := prabs-go

build:
	go build -o $(BIN) ./cmd/prabs

install:
	go install ./cmd/prabs

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

verify: build
	./$(BIN) verify .
