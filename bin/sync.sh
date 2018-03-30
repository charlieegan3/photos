#!/usr/bin/env bash

set -e

bucket="charlieegan3-instagram-archive"
folder="current"
path="$bucket/$folder"

aws s3 sync --acl public-read media s3://$path
gsutil rsync -a public-read media gs://$path
b2 sync --threads 4 media b2://$path

counts=()

counts+=($(aws s3 ls $path/ | wc -l))
counts+=($(gsutil ls gs://$path | wc -l))
counts+=($(b2 ls $bucket $folder | wc -l))

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
