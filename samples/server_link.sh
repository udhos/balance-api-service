#!/bin/bash

. ./helper.sh

curl -u "$AUTH" --data-binary "@server_link.yaml" -X POST -H "Accept: text/x-yaml" -H "Content-Type: text/x-yaml" "$URL"
