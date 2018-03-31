#!/usr/bin/env bash

set -e

if [ -z "$(git status --porcelain)" ]; then
  echo "Nothing to commit"
else
  git add .
  git -c user.name=automated -c user.email=git@charlieegan3.com commit -m "Add $(git status --porcelain | grep looted_json | wc | awk '{ print $1 }') images"
  git push origin master
fi
