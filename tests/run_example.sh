#!/bin/bash

cd $(git rev-parse --show-toplevel)

function err() {
    echo -e "\033[031m" $1 "\033[0m"
    exit $2
}

cmd="make run-example"
sec=30
timeout $sec $cmd >& /dev/null

status=$?
if [ $status -eq 124 ]; then
    err "$cmd timeouted: took over $sec s" $status
elif [ $status -ne 0 ]; then
    err "$cmd failed" $status
fi
