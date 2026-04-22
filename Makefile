.PHONY: build test lint tidy run clean

build:
	go build -o ccw ./cmd/ccw

test:
	go test ./... -race -coverprofile=coverage.out

lint:
	golangci-lint run

tidy:
	go mod tidy

run:
	go run ./cmd/ccw $(ARGS)

clean:
	rm -f ccw coverage.out
