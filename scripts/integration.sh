#!/usr/bin/env bash

GREEN="\033[00;32m"
RED="\033[00;31m"

status=0

function check() {
    if [ $status ] && [ $1 ]; then
        status=$1
    fi
}

function shouldfail() {
    if [ ! $1 ]; then
        status=1
    fi
}

apizza _config -e 2> /dev/null
shouldfail $?

apizza cart shouldnotbeincart &> /dev/null
shouldfail $?

if [ $TRAVIS_OS_NAME = "windows"]; then
    default_config="C:/Users/$USER/.apizza"
else
    default_config="$HOME/.apizza"
fi

configdir="$(apizza config -d)"
if [ $configdir != $default_config ]; then
    echo 'wrong config dir:' $configdir
    status=1
fi
unset configdir

configfile="$(apizza config -f)"
if [ $configfile != "$default_config/config.json" ]; then
    echo 'wrong config file:' $configfile
    status=1
fi
unset configfile


if [[ $status -eq 0 ]]; then
    echo -e "${GREEN}pass" "\033[0m"
else
    echo -e "${RED}failure\033[0m" $status
fi

exit $status