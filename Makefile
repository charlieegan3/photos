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
	test -n "$$(git status --porcelain)" && \
		git add . && \
		git -c user.name=automated -c user.email=git@charlieegan3.com commit -m "Add $(git status --porcelain | grep media | wc | awk '{ print $$1 }') images" && \
		git push origin master

notify:
	curl -s --form-string "token=$$PUSHOVER_TOKEN" \
			--form-string "user=$$PUSHOVER_USER" \
			--form-string "message=there was an error" \
			https://api.pushover.net/1/messages.json

# adhoc tasks
archive:
	zip -r "instagram_data_$$(date +"%Y-%m-%d-%H%M").zip" \
		looted_json completed_json locations media updated_at

calendar:
	./bin/calendar.sh
