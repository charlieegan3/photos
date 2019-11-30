#!/usr/bin/env bash

set -exo pipefail

source /etc/config/env

git clone --depth 1 https://charlieegan3:$GITHUB_TOKEN@github.com/charlieegan3/photos.git /app

make download
make sync
make save
