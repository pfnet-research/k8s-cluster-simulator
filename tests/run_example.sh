#!/bin/bash

cd $(git rev-parse --show-toplevel)

function err() {
    echo -e "\033[031m" $1 "\033[0m"
    exit $2
}

exe="bin/main"
sec=30
timeout $sec $exe >& /dev/null

status=$?
if [ $status -ne 0 ]; then
    err "$exe timeouted: took over $sec s" $status
fi
