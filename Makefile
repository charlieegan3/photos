GO_TEST_ARGS := ""
FILE_PATTERN := 'html\|plush\|go\|sql\|Makefile'
TAG := $(shell git rev-parse --short HEAD)
IMAGE := "eu.gcr.io/charlieegan3-photos/photos:$(TAG)"

dev_server:
	find . | grep $(FILE_PATTERN) | entr -r bash -c 'clear; go run main.go server --config config.dev.yaml'

new_migration:
	 migrate create -dir migrations -ext sql $(MIGRATION_NAME)

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

update_heroku_config:
	heroku config:set -a charlieegan3-photos CONFIG_STRING="$$(cat config.prod.yaml | base64)"

docker:
	docker build . -t $(IMAGE)
	docker push $(IMAGE)

