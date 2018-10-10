#!/bin/bash

gofmt -s -w ./f5-service
go tool fix ./f5-service
go tool vet ./f5-service

CGO_ENABLED=0 go test ./f5-service

CGO_ENABLED=0 go install ./f5-service
