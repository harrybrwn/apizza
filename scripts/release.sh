#!/usr/bin/env bash

set -e

oses=("linux" "windows" "darwin")
arch="amd64"


for os in ${oses[@]}; do
    ext="$(GOOS=$os go env GOEXE)"
    echo "GOOS=$os GOARCH=$arch go build -o apizza-$os-$arch$ext"
    GOOS=$os GOARCH=$arch go build -o release/apizza-$os-$arc$ext
done
