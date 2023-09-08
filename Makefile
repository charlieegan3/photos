GO_TEST_ARGS := ""
FILE_PATTERN := 'html\|css\|plush\|go\|sql\|Makefile'
TAG := $(shell git rev-parse --short HEAD)
IMAGE := "eu.gcr.io/charlieegan3-photos/photos:$(TAG)"

dev_server:
	find . | grep $(FILE_PATTERN) | entr -r bash -c 'clear; go run main.go server --config config.prod.yaml'

new_migration:
	 migrate create -dir internal/pkg/database/migrations -ext sql $(MIGRATION_NAME)

import:
	go run main.go import --config config.dev.yaml

test:
	go test ./... $(GO_TEST_ARGS)

test_watch:
	find . | grep $(FILE_PATTERN) | entr bash -c 'clear; make test'

test_db:
	docker run  -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres

local_bucket:
	python -m SimpleHTTPServer 8000

.PHONY: update_config
update_config:
	cat northflank-secret-template.json | \
		sed "s/DATA/$$(cat config.prod.yaml | base64)/g" > northflank-secret.json
	northflank update secret --projectId photos --secretId config --file northflank-secret.json

