#!/bin/bash

. ./helper.sh

set -x
curl -u "$AUTH" --data-binary "@server_delete.yaml" -X DELETE -H "Accept: text/x-yaml" -H "Content-Type: text/x-yaml" "$URL"

