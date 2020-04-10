#!/bin/bash

set -e

version="$(go version | sed -En 's/go version go(.*) .*/\1/p')"
if [ $version = "1.11" ]; then
    go list -f '{{join .Imports "\n"}}' ./... | \
        grep -P '(github\.com|gopkg\.in)/(?!harrybrwn)' \
        tr '\n' ' ' | \
        go get -u
fi

go build -o bin/test-apizza -ldflags "-X cmd.enableLog=false"
