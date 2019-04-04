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
Checks whether files have an appropriate license header.
"""

from pathlib import Path

from license_header import license_header, has_license_header

PROJECT_ROOT = Path(__file__).resolve().parents[1]


def main():
    ok = True

    ok &= check(
        PROJECT_ROOT.glob("*[!vendor]/**/*_k8s.go"),
        license_header("//", modification=True))
    ok &= check(
        [p for p in PROJECT_ROOT.glob(
            "*[!vendor]/**/*.go") if p.name[-7:] != "_k8s.go"],
        license_header("//"))
    ok &= check(PROJECT_ROOT.glob("*[!vendor]/**/*.py"), license_header("#"))
    ok &= check(PROJECT_ROOT.glob("*[!vendor]/**/*.sh"), license_header("#"))

    return 0 if ok else 1


def check(paths, license_header):
    ok = True

    for p in paths:
        if not p.exists():
            print("File", p.relative_to(PROJECT_ROOT), "does not exist")
            ok = False
            continue

        if not has_license_header(p, license_header):
            print(p.relative_to(PROJECT_ROOT), "is missing a license header")
            ok = False

    return ok


if __name__ == "__main__":
    exit(main())
