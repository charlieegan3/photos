dev_server:
	find . | grep 'plush\|go' | entr -r bash -c 'clear; go run main.go server --config config.dev.yaml'

test:
	go test ./...

test_watch:
	find . | grep 'plush\|go' | entr bash -c 'clear; make test'

test_db:
	docker run  -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres

local_bucket:
	python -m SimpleHTTPServer 8000
