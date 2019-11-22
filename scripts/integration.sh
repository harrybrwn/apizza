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

configdir="$(apizza config -d)"
if [ $configdir != "$HOME/.apizza" ]; then
    echo 'wrong config dir:' $configdir
    status=1
fi
unset configdir

configfile="$(apizza config -f)"
if [ $configfile != "$HOME/.apizza/config.json" ]; then
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