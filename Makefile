.PHONY: calendar

all_or_report:
	echo "Starting" > log 2>&1 && \
	make download >> log 2>&1 && \
	make sync >> log 2>&1 && \
	make save >> log 2>&1 \
	|| make notify

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

notify:
	curl -s --form-string "token=$$PUSHOVER_TOKEN" \
			--form-string "user=$$PUSHOVER_USER" \
			--form-string "message=$$(cat log)" \
			https://api.pushover.net/1/messages.json

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
	docker run --rm -it -e GOOGLE_PROJECT="$$GOOGLE_PROJECT" \
						-e GOOGLE_JSON="$$GOOGLE_JSON" \
						-e AWS_ACCESS_KEY_ID="$$AWS_ACCESS_KEY_ID" \
						-e AWS_REGION="$$AWS_REGION" \
						-e AWS_SECRET_ACCESS_KEY="$$AWS_SECRET_ACCESS_KEY" \
						-e B2_ACCOUNT_ID="$$B2_ACCOUNT_ID" \
						-e B2_ACCOUNT_KEY="$$B2_ACCOUNT_KEY" \
						-e GITHUB_TOKEN="$$GITHUB_TOKEN" \
						-e PUSHOVER_TOKEN="$$PUSHOVER_TOKEN" \
						-e PUSHOVER_USER="$$PUSHOVER_USER" \
						charlieegan3/photos:latest

hugo_server:
	cd site; hugo server; cd ..

hugo_build:
	cd site; hugo; cd ..

build_site:
	./hack/build_site.rb

vue_image:
	docker build -t charlieegan3/photos-vue -f Dockerfile.vue .

vue_serve:
	docker run -it -v $$(pwd)/frontend:/app -p 8080:8080 charlieegan3/photos-vue yarn serve

data_serve:
	cd output && ran -p 8000 -cors=true
