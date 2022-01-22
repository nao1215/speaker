#!/bin/bash -eu
# [Description]
#  This shell script is the installer for the command.
APP="speaker"
ROOT_DIR=$(git rev-parse --show-toplevel)

# warnMsg() output message in yellow to stdout.
# arg1: message.
function warnMsg() {
    local message="$1"
    echo -n -e "\033[33m\c"  # Escape sequence to make text color yellow
    echo "${message}"
    echo -n -e "\033[m\c"  # Escape sequence to restore font color
}

# errMsg() output message in red to stdout.
# arg1: message.
function errMsg() {
    local message="$1"
    echo -n -e "\033[31m\c"  # Escape sequence to make text color red
    echo "${message}" >&2
    echo -n -e "\033[m\c"   # Escape sequence to restore font color
}

# isRoot() check whether user is root or not.
function isRoot() {
    if [ ${EUID:-${UID}} != 0 ]; then
        echo "1"
        return
    fi
    echo "0"
}

function installSpeaker() {
    install -v -m 0755 -D speaker /usr/local/bin
}


IS_ROOT=$(isRoot)
if [ "$IS_ROOT" = "1" ]; then
    errMsg "[Usage]"
    errMsg " $ sudo ./installer.sh"
    exit 1
fi
warnMsg "[Start] Install."
installSpeaker
warnMsg "[Done]"