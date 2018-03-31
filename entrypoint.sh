#!/usr/bin/env bash

git clone --depth 1 https://charlieegan3:$GITHUB_TOKEN@github.com/charlieegan3/instagram-archive.git /app

make download && make sync && make save || make notify
