GO_TEST_ARGS := ""

dev_server:
	find . | grep 'plush\|go\|sql\|Makefile' | entr -r bash -c 'clear; go run main.go server --config config.dev.yaml'

import:
	go run main.go import --config config.dev.yaml

test:
	go test ./... $(GO_TEST_ARGS)

test_watch:
	find . | grep 'plush\|go\|sql\|Makefile' | entr bash -c 'clear; make test'

test_db:
	docker run  -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres

local_bucket:
	python -m SimpleHTTPServer 8000

update_heroku_config:
	heroku config:set -a charlieegan3-photos CONFIG_STRING="$(cat config.prod.yaml | base64)"
