#!/usr/bin/bash

go install
[[ "$TRAVIS_GO_VERSION" = "1.11" ]] && make install-deps