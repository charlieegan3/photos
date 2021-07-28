test_watch:
	find . | entr bash -c 'clear; go test ./...'

test_db:
	docker run  -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres
