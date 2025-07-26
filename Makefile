GO_TEST_ARGS := ""
FILE_PATTERN := 'html\|css\|plush\|go\|sql\|Makefile'
TAG := $(shell git rev-parse --short HEAD)
IMAGE := "eu.gcr.io/charlieegan3-photos/photos:$(TAG)"

dev:
	lsof -i :3000 | tail -n 1 | awk '{print $$2}' | xargs kill -9 || true
	find . | grep $(FILE_PATTERN) | entr -r bash -c 'clear; go run main.go server --config=config.dev.yaml'

new_migration:
	 migrate create -dir internal/pkg/database/migrations -ext sql $(MIGRATION_NAME)

import:
	go run main.go import --config config.dev.yaml

build:
	go build -o photos main.go

test:
	go test ./... $(GO_TEST_ARGS)

test_watch:
	find . | grep $(FILE_PATTERN) | entr bash -c 'clear; make GO_TEST_ARGS="$(GO_TEST_ARGS)" test'

test_db:
	docker run -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres

.PHONY: lint
lint:
	@echo "Running Go linter..."
	golangci-lint run ./...

local_bucket:
	python -m SimpleHTTPServer 8000

.PHONY: fmt
fmt:
	@echo "Running Go formatters..."
	@echo "Running goimports..."
	goimports -w .
	@echo "Running gofumpt..."
	gofumpt -w .
	@echo "Running go mod tidy..."
	go mod tidy
	@echo "Running dprint for JSON, Markdown, and HTML files..."
	dprint fmt
	@echo "Running treefmt for Nix files..."
	treefmt
	@echo "Formatting complete!"
