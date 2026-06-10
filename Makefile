.PHONY: build test vet lint install clean cover

BINARY := todo
PKG    := ./...

build:
	go build -o $(BINARY) ./cmd/todo

test:
	go test -race -cover $(PKG)

vet:
	go vet $(PKG)

lint: vet
	gofmt -l . | tee /dev/stderr | (! read)

install:
	go install ./cmd/todo

cover:
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f $(BINARY) coverage.out coverage.html
