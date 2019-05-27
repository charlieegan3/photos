#!/usr/bin/env bash

set -exo pipefail

container=$(docker create --workdir $(pwd) charlieegan3/photos-vue $@)

docker cp . $container:$(pwd)

docker start -ai $container

docker cp $container:$(pwd)/. .
