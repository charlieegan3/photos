#!/usr/bin/env bash

set -exo pipefail

if [ -z "$(git status --porcelain)" ]; then
  echo "Nothing to commit"
else
  git add .
  git -c user.name=automated \
      -c user.email=charlieegan3@users.noreply.github.com \
      commit -m "Add $(git status --porcelain | grep looted_json | wc | awk '{ print $1 }') images"
  git push origin master
fi
