#!/bin/bash

. ./helper.sh 

curl -u "$AUTH" --data-binary "@server_unlink.yaml" -X DELETE -H "Accept: text/x-yaml" -H "Content-Type: text/x-yaml" "$URL"

