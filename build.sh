#!/bin/bash

build() {
	local path="$1"

	gofmt -s -w $path
	go tool fix $path
	go tool vet $path

	CGO_ENABLED=0 go test $path
	CGO_ENABLED=0 go install $path
}

build ./examples/f5-api-client
build ./balance-service

