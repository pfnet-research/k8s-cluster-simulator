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

# Adds lincense header to files.
# FIXME: Reserve shebang

LICENSE_LINE='Licensed under the Apache License, Version 2.0 (the "License");'

LICENSE_HEADER=(
    'Copyright 2019 Preferred Networks, Inc.'
    ''
    'Licensed under the Apache License, Version 2.0 (the "License");'
    'you may not use this file except in compliance with the License.'
    'You may obtain a copy of the License at'
    ''
    '    http://www.apache.org/licenses/LICENSE-2.0'
    ''
    'Unless required by applicable law or agreed to in writing, software'
    'distributed under the License is distributed on an "AS IS" BASIS,'
    'WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.'
    'See the License for the specific language governing permissions and'
    'limitations under the License.'
)

cd $(git rev-parse --show-toplevel)

function err() {
    echo -e "\033[031m$1\033[0m"
    exit $2
}

files=$(git ls-files | grep -v vendor | grep -e ".go" -e ".sh" -e ".py")
for f in ${files[@]}; do
    if ! grep "$LICENSE_LINE" $f --quiet; then
        echo "Add license header to $f"

        b=$(basename $f)
        comment=""
        if [[ $b == *.go ]]; then
            comment="//"
        elif [[ $b == *.sh ]]; then
            comment="#"
        elif [[ $b == *.py ]]; then
            comment="#"
        else
            err "Unkown file extension" 1
        fi

        tmpfile=$(mktemp)
        for l in "${LICENSE_HEADER[@]}"; do
            if [ -z "$l" ]; then
                echo "${comment}" >> $tmpfile
            else
                echo "${comment} ${l}" >> $tmpfile
            fi
        done
        echo >> $tmpfile

        tmp2=$(mktemp)
        cat $tmpfile $f > $tmp2
        mv $tmp2 $f
    fi
done
