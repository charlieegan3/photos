#!/usr/bin/env bash

set -e

if [ -z "$(ls -A media)" ]; then
  echo "Nothing to sync"
  exit 0
fi

b2 authorize_account "$B2_ACCOUNT_ID" "$B2_ACCOUNT_KEY"

bucket="charlieegan3-instagram-archive"
folder="current"
path="$bucket/$folder"

aws s3 sync --size-only --acl public-read media s3://$path
gsutil -o "GSUtil:default_project_id=$GOOGLE_PROJECT" -h "Cache-Control: public, max-age=2629800" rsync -c -a public-read media gs://$path
b2 sync --compareVersions size --threads 4 media b2://$path

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
