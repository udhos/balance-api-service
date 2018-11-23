#!/bin/bash

. ./helper.sh

set -x
curl -u "$AUTH" -X GET -H "Accept: text/x-yaml" -H "Content-Type: text/x-yaml" "$URL"

