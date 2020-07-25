TAG=$(shell git rev-parse HEAD)

ifndef DOCKER_USERNAME
$(error DOCKER_USERNAME is not set)
endif
ifndef DOCKER_PASSWORD
$(error DOCKER_PASSWORD is not set)
endif

watch_test:
	find . | entr make test

test: clear
	go test ./...

clear:
	clear

install:
	go install

docker-login:
	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)

image: docker-login
	docker build . -t charlieegan3/photos:$(TAG) -t charlieegan3/photos:latest
	docker push charlieegan3/photos:$(TAG)
	docker push charlieegan3/photos:latest

get-binary:
	id=$$(docker create charlieegan3/photos:latest) && \
	docker cp $$id:/photos ./photos
