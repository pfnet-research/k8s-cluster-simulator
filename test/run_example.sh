#!/bin/bash

# Copyright 2019 Preferred Networks, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Tests that the example app finishes successfully without being stucked.

cd $(git rev-parse --show-toplevel)

function err() {
    echo -e "\033[031m$1\033[0m"
    exit $2
}

cmd="make run-example"
sec=30
timeout $sec $cmd >& /dev/null
rm kubesim.log kubesim-hr.log

status=$?
if [ $status -eq 124 ]; then
    err "'$cmd' timeouted: took over $sec s" $status
elif [ $status -ne 0 ]; then
    err "'$cmd' failed" $status
fi
