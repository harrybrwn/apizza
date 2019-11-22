#!/usr/bin/bash

set -e

go list -f '{{ join .Imports "\n" }}' ./... | \
		grep -P '^(github.com|gopkg.in)/.*' | \
		grep -v "`go list`" | \
		awk '{print}' ORS=' ' | \
		go get -u

go install -i github.com/harrybrwn/apizza