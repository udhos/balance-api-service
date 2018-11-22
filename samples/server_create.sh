#!/bin/bash

if [ -z "$URL" ]; then
	echo >&2 $0: missing env var URL=[$URL]
	exit 1
fi

curl -u admin:a10 --data-binary "@server_create.yaml" -X POST -H "Accept: text/x-yaml" -H "Content-Type: text/x-yaml" "$URL"

