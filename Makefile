TAG := $(shell tar -cf - . | md5sum | cut -f 1 -d " ")

.PHONY: calendar

download:
	./bin/1_fetch_raw.rb
	./bin/2_complete_json.rb
	./bin/3_save_locations.rb
	./bin/4_download_media.rb
	./bin/5_print_status.sh

sync:
	./bin/sync.sh

save:
	./bin/save.sh

# adhoc tasks
archive:
	zip -r "instagram_data_$$(date +"%Y-%m-%d-%H%M").zip" \
		looted_json completed_json locations media updated_at

calendar:
	./bin/calendar.sh

docker_image:
	docker build -t "charlieegan3/photos:$$(cat Dockerfile entrypoint.sh | shasum | awk '{ print $$1 }')" -t charlieegan3/photos:latest .
	docker push "charlieegan3/photos:$$(cat Dockerfile entrypoint.sh | shasum | awk '{ print $$1 }')"

docker_run:
	docker run --rm -it \
		--env-file=.envrc \
		-e ENV_PATH=/etc/config/env \
		-v $(PWD)/.envrc:/etc/config/env \
		-v $(PWD)/google.json:/etc/config/google \
		charlieegan3/photos:latest

###############################################
# frontend
###############################################

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
