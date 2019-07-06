#!/usr/bin/env sh

set -xeo pipefail

: ${USER?"USER must be set, e.g. charlieegan3"}
: ${REPO?"REPO must be set, e.g. mysite"}
: ${INTERVAL?"INTERVAL in seconds must be set, e.g. 3600"}
: ${DESTINATION?"DESTINATION must be set, e.g. /my/output/path"}

mkdir -p $DESTINATION

while [ 1 ]; do
	curl -LO https://github.com/$USER/$REPO/archive/master.zip
	unzip master.zip > /dev/null && rm master.zip

	cd $REPO-master && photos site debug && cd -

	rm -rf $DESTINATION/*
	mv $REPO-master/output/* $DESTINATION
	rm -rf $REPO-master

	echo "Sleeping for $INTERVAL seconds..." && sleep $INTERVAL
done
# docker build -t hugo-rebuilder . && docker run -it -e USER=charlieegan3 -e REPO=photos -e INTERVAL=30 -e DESTINATION=/output -e SOURCE_DIR="site" -e PRECOMMAND="./bin/build_site.rb" hugo-rebuilder
