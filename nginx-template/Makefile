TAG := $(shell tar -cf - . | md5sum | cut -f 1 -d " ")
PROJECT := photos-nginx-template

build:
	docker build -t charlieegan3/$(PROJECT):latest -t charlieegan3/$(PROJECT):${TAG} .

push: build
	docker push charlieegan3/$(PROJECT):latest
	docker push charlieegan3/$(PROJECT):${TAG}
