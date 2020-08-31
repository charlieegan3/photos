TAG=$(shell git rev-parse HEAD)

watch_test:
	find . | entr make test

test: clear
	go test ./...

clear:
	clear

install:
	go install

image:
	docker build . -t charlieegan3/photos:$(TAG) -t charlieegan3/photos:latest

get-binary: image
	id=$$(docker create charlieegan3/photos:latest) && \
	docker cp $$id:/photos ./photos
