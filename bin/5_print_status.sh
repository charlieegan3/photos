#!/usr/bin/env bash

files=( $(git status | grep media/ | awk '{ print $1 }') )

for file in "${files[@]}"; do
  size=$(ls -hl $file | awk '{ print $5 }')
  res=$(identify -format "%w * %h\n" $file)
  echo "$file:   $res    $size"
done

for dir in "looted_json" "completed_json" "media"; do
  count=$(ls $dir | wc | awk '{ print $1 }')
  echo -e "$count" "\t" $dir
done

profile="https://www.instagram.com/charlieegan3/"
echo $profile
curl --silent $profile | grep -oh --color "\S\+ Posts" | head -n 1
