#!/usr/bin/env bash

set -e
echo "" > coverage.txt

gotest -v ./... -coverprofile=coverage.txt -covermode=atomic
