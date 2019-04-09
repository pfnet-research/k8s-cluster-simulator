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
Adds a license header to files.
"""

from pathlib import Path
import subprocess

from license_header import license_header, has_license_header

PROJECT_ROOT = Path(subprocess.check_output(
    "git rev-parse --show-toplevel".split()).strip().decode("utf-8"))


def main(verbose=False):
    add(PROJECT_ROOT.glob("*[!vendor]/**/*_k8s.go"),
        license_header("//", True), verbose)
    add([p for p in PROJECT_ROOT.glob("*[!vendor]/**/*.go")
         if p.name[-7:] != "_k8s.go"], license_header("//"), verbose)
    add(PROJECT_ROOT.glob("*[!vendor]/**/*.py"), license_header("#"), verbose)
    add(PROJECT_ROOT.glob("*[!vendor]/**/*.sh"), license_header("#"), verbose)


def add(paths, license_header, verbose):
    for p in paths:
        if verbose:
            print("Checking", p.relative_to(PROJECT_ROOT))

        if has_license_header(p, license_header):
            continue

        print("Add license header to file", p.relative_to(PROJECT_ROOT))

        with p.open() as f:
            content = f.readlines()

            if content[0][:2] == "#!":  # reserve shebangs
                content_new = [content[0], "\n", license_header, "\n"]

                # put a newline after the license header only if necessary
                if content[1] != "\n":
                    content_new += ["\n"]
                content_new += content[1:]
            else:
                content_new = [license_header, "\n", "\n"] + content

        with p.open("w") as f:
            f.write(''.join(content_new))


if __name__ == "__main__":
    from argparse import ArgumentParser

    parser = ArgumentParser()
    parser.add_argument("-v", "--verbose", action="store_true")

    args = parser.parse_args()
    main(args.verbose)
