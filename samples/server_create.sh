#!/bin/bash

. ./helper.sh 

set -x
curl -u "$AUTH" --data-binary "@server_create.yaml" -X POST -H "Accept: text/x-yaml" -H "Content-Type: text/x-yaml" "$URL"

