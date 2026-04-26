.PHONY: bootstrap build test lint tidy run clean release-check release-snapshot release-clean coverage coverage-html

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
	rm -f ccw coverage.out coverage.html

release-check:
	goreleaser check

release-snapshot:
	HOMEBREW_TAP_GITHUB_TOKEN=dummy goreleaser release --snapshot --clean --skip=publish

release-clean:
	rm -rf dist/

coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "open coverage.html"

bootstrap:
	@command -v brew >/dev/null 2>&1 || { \
	  echo "❌ Homebrew is required. See https://brew.sh"; \
	  exit 1; \
	}
	brew bundle --file=Brewfile
	mise install
	lefthook install --force
	go mod download
	@echo "✅ bootstrap complete"
