#!/bin/sh

set -e

version="$(go version | sed -En 's/go version go(.*) .*/\1/p')"
if [ $version = "1.11" ]; then
    go list -f '{{join .Imports "\n"}}' ./... | \
        grep -P '(github\.com|gopkg\.in)/(?!harrybrwn)' \
        tr '\n' ' ' | \
        go get -u
fi

build_no="$(git describe --tags --abbrev=12)"
modpath="$(go list)"

version_flag="$modpath/cmd.version=$build_no"

if [ "$1" = "test" ]; then
    go build \
        -o bin/test-apizza \
        -ldflags "-X $modpath/cmd.enableLog=no -X ${version_flag}_test-build"
else
    go build -o bin/apizza -ldflags "-X $version_flag"
fi
