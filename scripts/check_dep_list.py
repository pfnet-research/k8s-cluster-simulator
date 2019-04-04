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

"""
Validates that DEPENDENCIES.md is synced with Gopkg.toml.
"""

from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[1]

WHITELIST_GOPKG = [  # may be missing in Gopkg.toml
    "k8s.io/api",
    "k8s.io/apimachinery",
]

WHITELIST_DEPLIST = [  # may be missing in DEPENDENCIES.md
]


def main():
    ok = True

    gopkg = deps_gopkg()

    deplist_names = []
    for d in deps_list():
        if len(d[1]) == 0:
            print('License of "{}" must not be empty.'.format(d[0], ))
            ok = False
            continue
        deplist_names.append(d[0])

    for g in gopkg:
        if g not in deplist_names and g not in WHITELIST_DEPLIST:
            print('"{}" is missing in DEPENDENCIES.md.'.format(g, ))
            ok = False

    for d in deplist_names:
        if d not in gopkg and d not in WHITELIST_GOPKG:
            print('DEPENDENCIES.md has "{}",'.format(d, ),
                  "despite that this package doesn't depend on it.")
            ok = False

    return ok


def deps_gopkg():
    import toml

    """
    Parses `Gopkg.toml` and returns a list of dependencies.
    """

    dep_names = []
    with (PROJECT_ROOT / "Gopkg.toml").open() as f:
        deps = toml.load(f)
        for c in deps["constraint"]:
            dep_names.append(c["name"])

    return dep_names


def deps_list():
    """
    Parses `DEPENDENCIES.md` and returns a list of (name, license) pairs.
    """

    import re

    RE_TABLE_HEADER = re.compile(r"^\| Name\s* \| License\s* \|$")
    RE_URL = re.compile(r"^\| \[(.+?)\]\(.+?\)\s* \| (.*)\s* \|$")
    RE_NOURL = re.compile(r"^\| (.+?)\s* \| (.*)\s* \|$")

    deps = []

    with (PROJECT_ROOT / "DEPENDENCIES.md").open() as f:
        # skip until header of deps table
        while not RE_TABLE_HEADER.match(f.readline()):
            pass
        f.readline()  # skip separation

        for l in f.readlines():
            matched = RE_URL.match(l)
            if matched:
                name = matched.group(1)
                license = matched.group(2)
                deps.append((name, license))
            else:
                matched = RE_NOURL.match(l)
                if matched:
                    name = matched.group(1)
                    license = matched.group(2)
                    deps.append((name, license))
                else:
                    break

    return deps


if __name__ == "__main__":
    exit(0 if main() else 1)
