#!/usr/bin/env bash

files=( $(git status | grep media/ | awk '{ print $1 }') )

for file in "${files[@]}"; do
  size=$(ls -hl $file | awk '{ print $5 }')
  res=$(identify -format "%w * %h\n" $file)
  echo "$file:   $res    $size"
done

counts=()

for dir in "looted_json" "completed_json" "media"; do
  count=$(ls $dir | wc | awk '{ print $1 }')
  echo -e "$count" "\t" $dir
  counts+=($count)
done

profile="https://www.instagram.com/charlieegan3/"
actual_count=$(curl --silent $profile | grep -oh --color "\S\+ Posts" | head -n 1 | sed s/[^0-9]//g)

echo "$actual_count posts found at $profile"

counts+=($actual_count)

if [ "${#counts[@]}" -gt 0 ] && [ $(printf "%s\000" "${counts[@]}" |
       LC_ALL=C sort -z -u |
       grep -z -c .) -ne 1 ] ; then
  echo "Missing images"
  exit 1
else
  echo "Counts match"
fi
