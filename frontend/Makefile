vue_image:
	docker build -t charlieegan3/photos-vue .

vue_install: vue_image
	docker run -t -v $$(pwd):/app charlieegan3/photos-vue yarn install

vue_serve: vue_install
	docker run -t --network="host" -v $$(pwd):/app -p 8080:8080 charlieegan3/photos-vue yarn serve

vue_build: vue_install
	docker run -t -v $$(pwd):/app charlieegan3/photos-vue yarn build

data_serve:
	photos site debug --output data && ran -p 8000 -cors=true
