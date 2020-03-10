#!/usr/bin/env bash

GREEN="\033[00;32m"
RED="\033[00;31m"

### Setup ################################################################

status=0
bin="$1"

if [ -z $bin ]; then
    bin="$(which apizza)"
fi

if [ ! -x $bin ]; then
    echo 'could not find binary'
    exit 1
fi

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

### Tests ###############################################################
echo "Running tests with $bin"

$bin _config -e 2> /dev/null
shouldfail $?

$bin cart shouldnotbeincart &> /dev/null
shouldfail $?

if [[ $TRAVIS_OS_NAME = "windows" ]]; then
    default_config="C:\\Users\\travis\\.config\\apizza"
    default_configfile="$default_config\\config.json"
else
    default_config="$HOME/.config/apizza"
    default_configfile="$default_config/config.json"
fi

$bin --help &> /dev/null

configdir="$($bin config -d)"
if [ $configdir != $default_config ]; then
    echo "wrong config dir... got: $configdir, want: $default_config"
    status=1
fi
unset configdir

configfile="$($bin config -f)"
if [ $configfile != "$default_configfile" ]; then
    echo "wrong config dir... got: $configfile, want: $default_configfile"
    status=1
fi
unset default_config
unset default_configfile
unset configfile

### Teardown #############################################################

if [[ $status -eq 0 ]]; then
    echo -e "${GREEN}Pass" "\033[0m"
else
    echo -e "${RED}Failure\033[0m:" $status
fi

exit $status
