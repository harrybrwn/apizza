#!/bin/bash

if [ "$TRAVIS_OS_NAME" = "windows" ]; then
    alias make='mingw23-make.exe'
fi
