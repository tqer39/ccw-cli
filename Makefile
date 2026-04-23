.PHONY: build test lint tidy run clean release-check release-snapshot release-clean

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

release-check:
	goreleaser check

release-snapshot:
	HOMEBREW_TAP_GITHUB_TOKEN=dummy goreleaser release --snapshot --clean --skip=publish

release-clean:
	rm -rf dist/
