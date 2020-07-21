watch_test:
	find . | entr make test

test: clear
	go test ./...

clear:
	clear

install:
	go install

binary:
	docker build -t photos:latest . && \
		id=$$(docker create photos:latest) && \
	docker cp $$id:/photos ./photos
