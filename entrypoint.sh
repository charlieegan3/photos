#!/usr/bin/env bash

git clone --depth 1 https://charlieegan3:$GITHUB_TOKEN@github.com/charlieegan3/photos.git /app

make all_or_report
