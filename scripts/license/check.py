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

import os
from pathlib import Path
import subprocess

from license_header import license_header, has_license_header

PROJECT_ROOT = subprocess.check_output(
    "git rev-parse --show-toplevel".split()).strip().decode("utf-8")
target_dirs = [os.path.join(PROJECT_ROOT, d) for d in os.listdir(PROJECT_ROOT)
               if d != "vendor" and os.path.isdir(os.path.join(PROJECT_ROOT, d))]


def main(verbose=False):
    ok = True

    ok &= check(Path(PROJECT_ROOT).glob("*.go"), license_header("//"), verbose)
    ok &= check(Path(PROJECT_ROOT).glob("*.py"), license_header("#"), verbose)
    ok &= check(Path(PROJECT_ROOT).glob("*.sh"), license_header("#"), verbose)

    for d in target_dirs:
        ok &= check(Path(d).glob("**/*_k8s.go"), license_header("//", modification=True), verbose)
        ok &= check([p for p in Path(d).glob("**/*.go")
            if p.name[-7:] != "_k8s.go"], license_header("//"), verbose)
        ok &= check(Path(d).glob("**/*.py"), license_header("#"), verbose)
        ok &= check(Path(d).glob("**/*.sh"), license_header("#"), verbose)

    return 0 if ok else 1


def check(paths, license_header, verbose):
    ok = True

    for p in paths:
        if verbose:
            print("Checking", p.relative_to(PROJECT_ROOT))

        if not p.exists():
            print("File", p.relative_to(PROJECT_ROOT), "does not exist")
            ok = False
            continue

        if not has_license_header(p, license_header):
            print(p.relative_to(PROJECT_ROOT), "is missing a license header")
            ok = False

    return ok


if __name__ == "__main__":
    from argparse import ArgumentParser

    parser = ArgumentParser()
    parser.add_argument("-v", "--verbose", action="store_true")

    args = parser.parse_args()
    exit(main(args.verbose))
