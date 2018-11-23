#!/bin/bash

if [ -z "$BASE_URL" ]; then
	BASE_URL=http://localhost:8080/v1
	echo >&2 $0: forcing empty env var BASE_URL="$BASE_URL"
fi

if [ -z "$NODE" ]; then
	NODE=1.1.1.1
	echo >&2 $0: forcing empty env var NODE="$NODE"
fi

if [ -z "$AUTH" ]; then
	AUTH=admin:a10
	echo >&2 $0: forcing empty env var AUTH="$AUTH"
fi

if [ -z "$URL" ]; then
	URL="$BASE_URL"/at2/node/"$NODE"/backend
	echo >&2 $0: forcing empty env var URL="$URL"
fi

cat <<__EOF__

BASE_URL and NODE are used only when URL is not set.

BASE_URL=$BASE_URL
NODE=$NODE
AUTH=$AUTH
URL=$URL

__EOF__
