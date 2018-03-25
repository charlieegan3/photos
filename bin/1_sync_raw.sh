#!/usr/bin/env bash

docker run -v "$(pwd)/looted_json:/out" -it python:alpine3.6 sh -c "pip install instaLooter && ls /out && instaLooter charlieegan3 /out -v -D -T {date}-{id} --time thisyear"

git checkout looted_json
