vue_image:
	docker build -t charlieegan3/photos-vue frontend

vue_install: vue_image
	docker run -it -v $$(pwd)/frontend:/app charlieegan3/photos-vue yarn install

vue_serve: vue_install
	docker run -it --network="host" -v $$(pwd)/frontend:/app -p 8080:8080 charlieegan3/photos-vue yarn serve

vue_build: vue_install
	docker run -it -v $$(pwd)/frontend:/app charlieegan3/photos-vue yarn build

data_serve:
	photos site debug --output data && ran -p 8000 -cors=true
