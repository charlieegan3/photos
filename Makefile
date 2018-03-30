all: download status save

download:
	./bin/1_fetch_raw.rb
	./bin/2_complete_json.rb
	./bin/3_save_locations.rb
	./bin/4_download_media.rb

status:
	./bin/5_print_status.sh

save:
	test -n "$$(git status --porcelain)" && \
		git add . && \
		git commit -m "Add $$(git status --porcelain | grep media | wc | awk '{ print $$1 }') images" && \
		git push origin master
