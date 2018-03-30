#!/usr/bin/env bash

set -e

if [ -z $YEAR ]; then
  YEAR=$(date +"%Y")
fi

mkdir -p calendar
rm calendar/* || true

echo $YEAR

for i in $(seq 1 12); do
  month=$(printf %02d $i)
  files=$(find media/ -type f -name "*.jpg" | grep $YEAR-$month | sort)
  count=$(echo "$files" | wc -l)
  tiles_wide=$(ruby -e "puts (Math.sqrt($count) * 1.3).round")

  if [ -z "$files" ]; then
    echo "Nothing found for $YEAR-$month"
  else
    echo "Generating $YEAR-$month..."
    montage $(echo $files | tr '\n' ' ') -tile "$tiles_wide"x -geometry 500x500 -border 10 -frame 0 -background white -bordercolor none calendar/$YEAR-$month.jpg
  fi
done
